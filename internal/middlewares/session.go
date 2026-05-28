package middlewares

import (
	"captcha_sweeper/internal/contexts"
	"captcha_sweeper/internal/cookies"
	db "captcha_sweeper/internal/database"
	"captcha_sweeper/internal/helpers"
	"captcha_sweeper/internal/servers"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tomasen/realip"
)

// SessionConfig holds dependencies for session middleware.
// It gets (or creates) a session_id and stores it in the request context.
// The session_id is also persisted in a signed cookie.
type SessionConfig struct {
	Queries          *db.Queries
	ErrorHelper      *servers.Errors
	CookieName       string
	CookieSigningKey string
}

func LoadOrCreateSession(cfg SessionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check the cookie name and signing key are set
			if cfg.CookieName == "" || cfg.CookieSigningKey == "" {
				// error: cookie name or signing key not set
				cfg.ErrorHelper.ServerError(w, r, errors.New("cookie name or signing key not set"))
				return
			}

			// get Cookie, ip and user agent
			cookieValue, err := cookies.ReadSigned(r, cfg.CookieName, cfg.CookieSigningKey)
			if err != nil {
				// no cookie or invalid signature. We'll create/get a DB session.
				cookieValue = ""
			}
			ip := realip.FromRequest(r)
			ua := r.UserAgent()

			sessionID, err := helpers.NewPgUUID(&cookieValue)
			if err != nil {
				sessionID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
			}

			// Always ensure a DB session exists (your GetOrCreateSession does upsert by ip+ua).
			session, err := cfg.Queries.GetOrCreateSession(r.Context(), db.GetOrCreateSessionParams{
				SessionID: sessionID,
				IpAddress: ip,
				UserAgent: ua,
			})
			if err != nil {
				cfg.ErrorHelper.ServerError(w, r, err)
				return
			}

			// Write cookie if missing or doesn't match the current session id.
			// (Even if it matches, writing again is harmless but avoids extra Set-Cookie noise.)
			if cookieValue == "" || cookieValue != session.SessionID.String() {
				err = cookies.WriteSigned(w, http.Cookie{
					Name:     cfg.CookieName,
					Value:    session.SessionID.String(),
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				}, cfg.CookieSigningKey)
				if err != nil {
					cfg.ErrorHelper.ServerError(w, r, err)
					return
				}
			}

			r = contexts.ContextSetAuthenticatedSession(r, session)
			next.ServeHTTP(w, r)
		})
	}
}
