package middlewares

import (
	"captcha_sweeper/internal/contexts"
	db "captcha_sweeper/internal/database"
	"captcha_sweeper/internal/servers"
	"net/http"
	"strings"
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

func isBotUserAgent(ua string) bool {
	u := strings.ToLower(ua)

	// Known bots
	if strings.Contains(u, "googlebot") ||
		strings.Contains(u, "bingbot") ||
		strings.Contains(u, "duckduckbot") ||
		strings.Contains(u, "baiduspider") ||
		strings.Contains(u, "yandex") ||
		strings.Contains(u, "sogou") ||
		strings.Contains(u, "slurp") {
		return true
	}

	// Common scraper/http clients
	if strings.Contains(u, "python-urllib") ||
		strings.Contains(u, "python-requests") ||
		strings.Contains(u, "aiohttp") ||
		strings.Contains(u, "scrapy") ||
		strings.Contains(u, "okhttp") ||
		strings.Contains(u, "curl") ||
		strings.Contains(u, "wget") {
		return true
	}

	// Generic fallback
	if strings.Contains(u, " bot") || strings.HasSuffix(u, "bot") || strings.Contains(u, "crawler") || strings.Contains(u, "spider") {
		return true
	}

	return false
}
