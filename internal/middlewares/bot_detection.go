package middlewares

import (
	"captcha_sweeper/internal/contexts"
	db "captcha_sweeper/internal/database"
	"captcha_sweeper/internal/servers"
	"net/http"
)

type BotDetectionConfig struct {
	Queries     *db.Queries
	ErrorHelper *servers.Errors
}

// BotDetection should be used AFTER the session middleware.
// It updates the session heuristics in the DB (request_count, avg_time_between_requests, is_bot)
// using the sqlc-generated UpdateSession query, and refreshes the session stored in request context.
func BotDetection(cfg BotDetectionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, exist := contexts.ContextGetAuthenticatedSession(r)
			if !exist {
				next.ServeHTTP(w, r)
				return
			}

			updated, err := cfg.Queries.UpdateSession(r.Context(), session.SessionID)
			if err != nil {
				cfg.ErrorHelper.ServerError(w, r, err)
				return
			}

			r = contexts.ContextSetAuthenticatedSession(r, updated)
			next.ServeHTTP(w, r)
		})
	}
}
