package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/cameronsralla/culdechat/server/internal/httpx"
	"github.com/cameronsralla/culdechat/server/internal/middleware"
)

// Identity is the authenticated caller, derived from the access token only.
// Handlers that need fresh user state (status, admin flag) load the row.
type Identity struct {
	UserID  uuid.UUID
	IsAdmin bool
}

type ctxKey struct{}

func init() {
	middleware.SetUserIDResolver(func(ctx context.Context) string {
		if id, ok := IdentityFrom(ctx); ok {
			return id.UserID.String()
		}
		return ""
	})
}

// IdentityFrom returns the caller identity if the request was authenticated.
func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// MustIdentity is for handlers behind RequireUser; it panics if missing.
func MustIdentity(ctx context.Context) Identity {
	id, ok := IdentityFrom(ctx)
	if !ok {
		panic("auth: handler reached without identity; missing RequireUser")
	}
	return id
}

// Authenticate parses a bearer token if present and attaches the identity.
// It never rejects; guards do that.
func (t *Tokens) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := bearer(r)
		if raw == "" {
			next.ServeHTTP(w, r)
			return
		}
		claims, err := t.Parse(raw)
		if err != nil {
			httpx.Fail(w, r, httpx.ErrUnauthorized.WithMessage("invalid or expired token"))
			return
		}
		uid, _ := uuid.Parse(claims.Subject)
		ctx := context.WithValue(r.Context(), ctxKey{}, Identity{UserID: uid, IsAdmin: claims.IsAdmin})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// RequireUser rejects unauthenticated requests.
func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := IdentityFrom(r.Context()); !ok {
			httpx.Fail(w, r, httpx.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin rejects non-admin requests.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := IdentityFrom(r.Context())
		if !ok {
			httpx.Fail(w, r, httpx.ErrUnauthorized)
			return
		}
		if !id.IsAdmin {
			httpx.Fail(w, r, httpx.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
