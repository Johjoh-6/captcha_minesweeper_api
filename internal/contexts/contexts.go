package contexts

import (
	db "captcha_sweeper/internal/database"
	"context"
	"net/http"
)

type contextKey string

const (
	sessionContextKey = contextKey("session")
)

func ContextSetAuthenticatedSession(r *http.Request, session db.Session) *http.Request {
	ctx := context.WithValue(r.Context(), sessionContextKey, session)
	return r.WithContext(ctx)
}

func ContextGetAuthenticatedSession(r *http.Request) (db.Session, bool) {
	session, ok := r.Context().Value(sessionContextKey).(db.Session)
	return session, ok
}
