package middlewares

import (
	"context"
	"net/http"
	"time"
)

const defaultRequestTimeout = 5 * time.Second

// Timeout returns a middleware that enforces a request timeout.
func Timeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), defaultRequestTimeout)
		defer cancel()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
