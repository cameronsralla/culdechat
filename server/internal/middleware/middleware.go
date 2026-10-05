// Package middleware is the request pipeline applied to every route:
// RequestID -> RealIP -> Logger -> Recover -> Timeout -> BodyLimit ->
// SecurityHeaders -> CORS -> RateLimit -> (Auth, per route group) -> handler.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/cameronsralla/culdechat/server/internal/httpx"
	applog "github.com/cameronsralla/culdechat/server/internal/log"
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
)

const requestIDHeader = "X-Request-Id"

// RequestID assigns a random id to each request and echoes it in the response.
// Client-provided ids are ignored (they are untrusted).
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b [12]byte
		_, _ = rand.Read(b[:])
		id := hex.EncodeToString(b[:])
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

// RequestIDFrom returns the request id set by RequestID.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// RealIP rewrites r.RemoteAddr from X-Forwarded-For / X-Real-IP only when the
// immediate peer is a trusted proxy. With no trusted proxies configured the
// headers are ignored, which is the safe default.
func RealIP(trusted []string) func(http.Handler) http.Handler {
	nets := parseCIDRs(trusted)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			if peer := net.ParseIP(host); peer != nil && contains(nets, peer) {
				if ip := forwardedIP(r); ip != "" {
					r.RemoteAddr = net.JoinHostPort(ip, "0")
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func forwardedIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		// Right-most entry was appended by the trusted proxy.
		candidate := strings.TrimSpace(parts[len(parts)-1])
		if net.ParseIP(candidate) != nil {
			return candidate
		}
	}
	if xr := strings.TrimSpace(r.Header.Get("X-Real-Ip")); net.ParseIP(xr) != nil {
		return xr
	}
	return ""
}

// ClientIP returns the host part of RemoteAddr (post-RealIP).
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func parseCIDRs(list []string) []*net.IPNet {
	var out []*net.IPNet
	for _, s := range list {
		if !strings.Contains(s, "/") {
			if strings.Contains(s, ":") {
				s += "/128"
			} else {
				s += "/32"
			}
		}
		if _, n, err := net.ParseCIDR(s); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func contains(nets []*net.IPNet, ip net.IP) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// Logger attaches a request-scoped logger and emits one line per request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		l := applog.From(r.Context()).With("request_id", RequestIDFrom(r.Context()))
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		r = r.WithContext(applog.With(r.Context(), l))

		next.ServeHTTP(ww, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"dur_ms", time.Since(start).Milliseconds(),
			"ip", ClientIP(r),
		}
		if uid := userIDFromContext(r.Context()); uid != "" {
			attrs = append(attrs, "user_id", uid)
		}
		switch {
		case ww.Status() >= 500:
			l.Error("http", attrs...)
		case ww.Status() >= 400:
			l.Warn("http", attrs...)
		default:
			l.Info("http", attrs...)
		}
	})
}

// userIDFromContext is wired by the auth package to avoid an import cycle.
var userIDFromContext = func(context.Context) string { return "" }

// SetUserIDResolver lets the auth package expose the current user id to the logger.
func SetUserIDResolver(f func(context.Context) string) { userIDFromContext = f }

// Recover converts panics into 500s with a stack trace in the log.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				applog.From(r.Context()).Error("panic", "err", rec, "stack", string(debug.Stack()))
				httpx.JSON(w, http.StatusInternalServerError, map[string]any{
					"error": map[string]string{"code": "internal", "message": "something went wrong"},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Timeout cancels the request context after d. Handlers must respect ctx.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// BodyLimit caps request bodies. Upload routes should wrap with a larger limit.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sets API-appropriate hardening headers. HSTS is set only
// when prod is true; Caddy terminates TLS and the browser must only see it over https.
func SecurityHeaders(prod bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Cache-Control", "no-store")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			if prod {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORS implements an exact-match origin allow-list. Credentials are not
// allowed (the client sends bearer tokens, not cookies).
func CORS(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[strings.TrimRight(o, "/")] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := allowed[origin]; !ok {
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			h := w.Header()
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Expose-Headers", "Content-Length, "+requestIDHeader)
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				h.Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
