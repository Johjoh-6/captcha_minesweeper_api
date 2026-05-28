package middlewares

import (
	"captcha_sweeper/internal/servers"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type BasicAuthConfig struct {
	Username       string
	HashedPassword string
	ErrorHelper    *servers.Errors
}

func RequireBasicAuth(cfg BasicAuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, plaintextPassword, ok := r.BasicAuth()
			if !ok || cfg.Username != username {
				basicAuthRequired(w)
				return
			}
			err := bcrypt.CompareHashAndPassword([]byte(cfg.HashedPassword), []byte(plaintextPassword))
			switch {
			case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
				basicAuthRequired(w)
				return
			case err != nil:
				cfg.ErrorHelper.ServerError(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func basicAuthRequired(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	w.WriteHeader(http.StatusUnauthorized)
}
