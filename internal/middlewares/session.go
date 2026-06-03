package middlewares

import (
	"captcha_sweeper/internal/contexts"
	"captcha_sweeper/internal/cookies"
	db "captcha_sweeper/internal/database"
	"captcha_sweeper/internal/helpers"
	"captcha_sweeper/internal/servers"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tomasen/realip"
)

type SessionConfig struct {
	IdentifierMode    servers.IdentifierMode
	Queries           *db.Queries
	ErrorHelper       *servers.Errors
	CookieName        string
	CookieSigningKey  string
	JWTSecretKey      []byte
	JWTExpiryDuration time.Duration
	SessionHeaderName string
}

type SessionClaims struct {
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// --- Helpers ---

func parseSessionID(value string) (pgtype.UUID, error) {
	if value == "" {
		return pgtype.UUID{Bytes: uuid.New(), Valid: true}, nil
	}
	return helpers.NewPgUUID(&value)
}

func writeSessionCookie(w http.ResponseWriter, cfg SessionConfig, sessionID pgtype.UUID) error {
	return cookies.WriteSigned(w, http.Cookie{
		Name:     "__Secure-" + cfg.CookieName,
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // for test http.SameSiteNoneMode
	}, cfg.CookieSigningKey)
}

func extractSessionIDFromJWT(cfg SessionConfig, r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}

	claims := &SessionClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return cfg.JWTSecretKey, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid JWT: %v", err)
	}

	return claims.SessionID, nil
}

func CreateJWT(sessionID pgtype.UUID, cfg SessionConfig) (string, error) {
	claims := SessionClaims{
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTExpiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.JWTSecretKey)
}

// --- Main Middleware ---
func LoadOrCreateSession(cfg SessionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sessionID pgtype.UUID
			ip := realip.FromRequest(r)
			ua := r.UserAgent()

			switch cfg.IdentifierMode {
			case servers.IdentifierModeCookie:
				if cfg.CookieName == "" || cfg.CookieSigningKey == "" {
					cfg.ErrorHelper.ServerError(w, r, errors.New("cookie name or signing key not set"))
					return
				}

				cookieValue, err := cookies.ReadSigned(r, "__Secure-"+cfg.CookieName, cfg.CookieSigningKey)
				if err != nil {
					cookieValue = ""
				}
				sessionID, err = parseSessionID(cookieValue)
				if err != nil {
					cfg.ErrorHelper.ServerError(w, r, err)
					return
				}

			case servers.IdentifierModeJWT:
				if len(cfg.JWTSecretKey) == 0 {
					cfg.ErrorHelper.ServerError(w, r, errors.New("JWT secret key not set"))
					return
				}

				extractedID, err := extractSessionIDFromJWT(cfg, r)
				if err != nil {
					extractedID = ""
				}
				sessionID, err = parseSessionID(extractedID)
				if err != nil {
					cfg.ErrorHelper.ServerError(w, r, err)
					return
				}

			case servers.IdentifierModeSession:
				if cfg.SessionHeaderName == "" {
					cfg.ErrorHelper.ServerError(w, r, errors.New("session header name not set"))
					return
				}

				sessionIDStr := r.Header.Get(cfg.SessionHeaderName)
				var err error
				sessionID, err = parseSessionID(sessionIDStr)
				if err != nil {
					cfg.ErrorHelper.ServerError(w, r, err)
					return
				}

			default:
				cfg.ErrorHelper.ServerError(w, r, fmt.Errorf("unsupported identifier mode: %v", cfg.IdentifierMode))
				return
			}

			// Get or create session in database
			session, err := cfg.Queries.GetOrCreateSession(r.Context(), db.GetOrCreateSessionParams{
				SessionID: sessionID,
				IpAddress: ip,
				UserAgent: ua,
			})
			if err != nil {
				cfg.ErrorHelper.ServerError(w, r, err)
				return
			}

			// Send back session ID
			switch cfg.IdentifierMode {
			case servers.IdentifierModeCookie:
				if err := writeSessionCookie(w, cfg, session.SessionID); err != nil {
					cfg.ErrorHelper.ServerError(w, r, err)
					return
				}
			case servers.IdentifierModeJWT:
				tokenString, err := CreateJWT(session.SessionID, cfg)
				if err != nil {
					cfg.ErrorHelper.ServerError(w, r, err)
					return
				}
				w.Header().Set("Authorization", "Bearer "+tokenString)
			case servers.IdentifierModeSession:
				w.Header().Set(cfg.SessionHeaderName, session.SessionID.String())
			}

			r = contexts.ContextSetAuthenticatedSession(r, session)
			next.ServeHTTP(w, r)
		})
	}
}
