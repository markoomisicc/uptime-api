package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/markoomisicc/uptime-api/internal/store"
)

const (
	sessionCookieName = "session"
	sessionDuration   = 30 * 24 * time.Hour
)

type contextKey string

const userContextKey contextKey = "user"

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) createSession(r *http.Request, userID uuid.UUID) (string, time.Time, error) {
	token := rand.Text()
	expiresAt := time.Now().Add(sessionDuration)

	err := s.queries.CreateSession(r.Context(), store.CreateSessionParams{
		UserID:    userID,
		TokenHash: hashToken(token),
		UserAgent: r.UserAgent(),
		IpAddress: clientIP(r),
		ExpiresAt: expiresAt,
	})

	return token, expiresAt, err
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   s.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication_required", "Authentication required")
			return
		}

		row, err := s.queries.GetUserBySessionToken(r.Context(), hashToken(cookie.Value))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusUnauthorized, "authentication_required", "Authentication required")
				return
			}
			writeInternalError(w, "get session", err)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, row.User)
		next(w, r.WithContext(ctx))
	}
}

func userFromContext(ctx context.Context) store.User {
	return ctx.Value(userContextKey).(store.User)
}
