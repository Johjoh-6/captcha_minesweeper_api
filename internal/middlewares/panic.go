package middlewares

import (
	"captcha_sweeper/internal/servers"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecoverPanic returns a middleware that recovers from panics.
func RecoverPanic(logger *slog.Logger, errHelper *servers.Errors) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if pv := recover(); pv != nil {
					logger.Error("panic recovered", "panic", pv, "trace", string(debug.Stack()))
					errHelper.ServerError(w, r, fmt.Errorf("%v", pv))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
