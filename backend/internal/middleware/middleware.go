// Package middleware contains HTTP middleware for the API.
package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/akeelnazir/osint-app/backend/internal/auth"
	"github.com/akeelnazir/osint-app/backend/internal/ent"
	"github.com/akeelnazir/osint-app/backend/internal/httperr"
	"golang.org/x/time/rate"
)

type ctxKey string

const (
	// CtxUser is the context key holding the authenticated *ent.User.
	CtxUser ctxKey = "user"
	// CtxUserID is the context key holding the authenticated user's int ID.
	CtxUserID ctxKey = "user_id"
	// CtxRole is the context key holding the authenticated user's role string.
	CtxRole ctxKey = "role"
)

// RequireAuth middleware validates the Bearer access token and loads the user
// from the DB, attaching it to the request context.
func RequireAuth(authSvc *auth.Service, client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httperr.Write(w, httperr.Unauthorized("missing or malformed Authorization header"))
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := authSvc.VerifyAccessToken(tokenStr)
			if err != nil {
				httperr.Write(w, httperr.Unauthorized("invalid or expired access token"))
				return
			}
			user, err := client.User.Get(r.Context(), claims.UserID)
			if err != nil {
				httperr.Write(w, httperr.Unauthorized("user no longer exists"))
				return
			}
			ctx := context.WithValue(r.Context(), CtxUser, user)
			ctx = context.WithValue(ctx, CtxUserID, user.ID)
			ctx = context.WithValue(ctx, CtxRole, string(user.Role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that allows only the listed roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(CtxRole).(string)
			if _, ok := allowed[role]; !ok {
				httperr.Write(w, httperr.Forbidden("insufficient role privileges"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserFromContext extracts the authenticated user, if present.
func UserFromContext(r *http.Request) (*ent.User, bool) {
	u, ok := r.Context().Value(CtxUser).(*ent.User)
	return u, ok
}

// UserIDFromContext extracts the authenticated user ID, if present.
func UserIDFromContext(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(CtxUserID).(int)
	return id, ok
}

// RateLimit applies a per-IP token-bucket rate limiter.
func RateLimit(rps int) func(http.Handler) http.Handler {
	var (
		mu       sync.Mutex
		buckets  = make(map[string]*rate.Limiter)
	)
	get := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		l, ok := buckets[ip]
		if !ok {
			l = rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), rps)
			buckets[ip] = l
		}
		return l
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.TrimSpace(strings.Split(fwd, ",")[0])
			}
			if !get(ip).Allow() {
				httperr.Write(w, httperr.New("rate_limited", "too many requests", http.StatusTooManyRequests))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeaders sets common security response headers.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// Recover catches panics and returns 500.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				httperr.Write(w, httperr.Internal("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
