package api

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type Middleware struct {
	auth AuthService
	l    *slog.Logger
}

func NewMiddleware(auth AuthService, l *slog.Logger) *Middleware {
	return &Middleware{
		auth: auth,
		l:    l,
	}
}

func (mw *Middleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.NewString()

		logger := mw.l.With("request_id", reqID)
		logger.Info("Incoming Request", "method", r.Method, "path", r.URL.Path)

		ctx := r.Context()
		ctx = context.WithValue(ctx, "logger", logger)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (mw *Middleware) Auth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cookie, err := r.Cookie("session_id")
		if err != nil {
			sendError(ctx, w, err)
			return
		}

		user, err := mw.auth.UserBySessionID(context.Background(), cookie.Value)
		if err != nil {
			sendError(ctx, w, err)
			return
		}

		ctx = context.WithValue(ctx, "user", user)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
