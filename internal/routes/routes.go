package routes

import (
	"captcha_sweeper/internal/handlers"
	"captcha_sweeper/internal/middlewares"
	"captcha_sweeper/internal/servers"
	"expvar"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func NewRouter(app *servers.Application) http.Handler {
	mux := mux.NewRouter()

	mux.NotFoundHandler = http.HandlerFunc(app.Errors.NotFound)
	mux.MethodNotAllowedHandler = http.HandlerFunc(app.Errors.MethodNotAllowed)

	// Initialize middleware with dependencies
	sessionMiddleware := middlewares.LoadOrCreateSession(middlewares.SessionConfig{
		Queries:          app.Store.Queries,
		ErrorHelper:      app.Errors,
		CookieName:       app.Config.Cookie.Name,
		CookieSigningKey: app.Config.Cookie.SecretKey,
	})
	botMiddleware := middlewares.BotDetection(middlewares.BotDetectionConfig{
		Queries:     app.Store.Queries,
		ErrorHelper: app.Errors,
	})
	rateLimiterStore := middlewares.NewRateLimiterStore()
	rateLimitMiddleware := middlewares.RateLimit(middlewares.RateLimitConfig{
		Store:       rateLimiterStore,
		ErrorHelper: app.Errors,
	})

	h := handlers.NewHandlers(app.Store.Queries, app.Logger, app.Errors, app.Config)

	// Apply middleware
	mux.Use(middlewares.RecoverPanic(app.Logger, app.Errors))
	mux.Use(middlewares.LogAccess(app.Logger))
	mux.Use(middlewares.Timeout)

	// CORS
	c := cors.New(cors.Options{
		// Reflect the request Origin so credentialed requests from browsers work.
		// Consider restricting this to known origins in production.
		AllowedOrigins: []string{
			"https://johjoh-6.github.io", // GitHub Pages
			"http://localhost:5173",      // Local dev
			"http://127.0.0.1:5173",
		},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Origin", "X-Requested-With"},
		AllowCredentials: true,
	})

	api := mux.PathPrefix("/v1").Subrouter()
	api.HandleFunc("/status", h.Status).Methods("GET")

	captchaRoute := api.PathPrefix("/captcha").Subrouter()
	captchaRoute.Use(sessionMiddleware, botMiddleware, rateLimitMiddleware)
	captchaRoute.HandleFunc("", h.Captcha.GetCaptcha).Methods("GET")
	captchaRoute.HandleFunc("/new", h.Captcha.NewCaptcha).Methods("POST")
	captchaRoute.HandleFunc("/move", h.Captcha.PlayCaptcha).Methods("POST")

	restrictedRoutes := mux.NewRoute().Subrouter()
	restrictedRoutes.Use(middlewares.RequireBasicAuth(middlewares.BasicAuthConfig{
		Username:       app.Config.BasicAuth.Username,
		HashedPassword: app.Config.BasicAuth.HashedPassword,
		ErrorHelper:    app.Errors,
	}))
	restrictedRoutes.Handle("/metrics", expvar.Handler()).Methods("GET")

	return c.Handler(mux)
}
