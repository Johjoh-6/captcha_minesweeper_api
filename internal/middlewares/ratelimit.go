package middlewares

import (
	"captcha_sweeper/internal/contexts"
	"captcha_sweeper/internal/response"
	"captcha_sweeper/internal/servers"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu             sync.Mutex
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
}

// NewRateLimiter creates a new token bucket rate limiter.
// requestsPerPeriod: number of allowed requests
// periodSeconds: time period in seconds
func NewRateLimiter(requestsPerPeriod int, periodSeconds int) *RateLimiter {
	return &RateLimiter{
		tokens:         float64(requestsPerPeriod),
		maxTokens:      float64(requestsPerPeriod),
		refillRate:     float64(requestsPerPeriod) / float64(periodSeconds),
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request should be allowed. Returns true if allowed, false if rate limited.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()
	rl.lastRefillTime = now

	// Refill tokens based on elapsed time
	rl.tokens = rl.tokens + (elapsed * rl.refillRate)
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	// Check if we have tokens available
	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

// RateLimiterStore holds rate limiters per session using sync.Map for thread-safe storage.
type RateLimiterStore struct {
	limiters sync.Map // map[string]*RateLimiter (keyed by session ID)
}

// NewRateLimiterStore creates a new RateLimiterStore instance.
func NewRateLimiterStore() *RateLimiterStore {
	return &RateLimiterStore{}
}

// RateLimitRule defines per-session limits for a specific bucket.
type RateLimitRule struct {
	Key           string
	Requests      int
	PeriodSeconds int
}

// GetOrCreateLimiter returns or creates a rate limiter for the given key.
func (s *RateLimiterStore) GetOrCreateLimiter(key string, rule RateLimitRule) *RateLimiter {
	limiter, ok := s.limiters.Load(key)
	if ok {
		return limiter.(*RateLimiter)
	}

	newLimiter := NewRateLimiter(rule.Requests, rule.PeriodSeconds)
	// Store and return (another goroutine may have stored the same key in the meantime)
	actual, _ := s.limiters.LoadOrStore(key, newLimiter)
	return actual.(*RateLimiter)
}

// RateLimitConfig holds dependencies for rate limiting middleware.
type RateLimitConfig struct {
	Store       *RateLimiterStore
	ErrorHelper *servers.Errors
}

func rateLimitRuleForRequest(r *http.Request, isBot bool) RateLimitRule {
	const periodSeconds = 60
	path := r.URL.Path

	if strings.HasSuffix(path, "/captcha/move") {
		if isBot {
			return RateLimitRule{Key: "captcha_move_bot", Requests: 60, PeriodSeconds: periodSeconds}
		}
		return RateLimitRule{Key: "captcha_move_user", Requests: 240, PeriodSeconds: periodSeconds}
	}

	if strings.HasSuffix(path, "/captcha/new") {
		if isBot {
			return RateLimitRule{Key: "captcha_new_bot", Requests: 10, PeriodSeconds: periodSeconds}
		}
		return RateLimitRule{Key: "captcha_new_user", Requests: 30, PeriodSeconds: periodSeconds}
	}

	if isBot {
		return RateLimitRule{Key: "captcha_other_bot", Requests: 20, PeriodSeconds: periodSeconds}
	}
	return RateLimitRule{Key: "captcha_other_user", Requests: 60, PeriodSeconds: periodSeconds}
}

// RateLimit returns middleware that enforces rate limits per session.
// Should be used AFTER session and bot detection middleware.
// Returns 429 Too Many Requests when limit is exceeded.
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, exist := contexts.ContextGetAuthenticatedSession(r)
			if !exist {
				next.ServeHTTP(w, r)
				return
			}

			sessionID := session.SessionID.String()
			isBot := session.IsBot.Valid && session.IsBot.Bool

			rule := rateLimitRuleForRequest(r, isBot)
			limiterKey := sessionID + ":" + rule.Key
			limiter := cfg.Store.GetOrCreateLimiter(limiterKey, rule)
			if !limiter.Allow() {
				// Rate limit exceeded - return 429 Too Many Requests
				w.Header().Set("Retry-After", "60")
				if err := response.JSON(w, http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
				}); err != nil {
					cfg.ErrorHelper.ServerError(w, r, err)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
