package middlewares

import (
	"captcha_sweeper/internal/response"
	"log/slog"
	"net/http"

	"github.com/tomasen/realip"
)

func LogAccess(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw := response.NewMetricsResponseWriter(w)
			next.ServeHTTP(mw, r)

			var (
				ip     = realip.FromRequest(r)
				method = r.Method
				url    = r.URL.String()
				proto  = r.Proto
			)

			userAttrs := slog.Group("user", "ip", ip)
			requestAttrs := slog.Group("request", "method", method, "url", url, "proto", proto)
			responseAttrs := slog.Group("response", "status", mw.StatusCode, "size", mw.BytesCount)

			logger.Info("access", userAttrs, requestAttrs, responseAttrs)
		})
	}
}
