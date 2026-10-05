package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/cameronsralla/culdechat/server/internal/httpx"
)

// Limiter is a keyed token bucket. Keys are usually client IP (anonymous
// routes) or user id (authenticated routes). Idle entries are swept.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	rate    rate.Limit
	burst   int
	enabled bool
}

type entry struct {
	lim  *rate.Limiter
	seen time.Time
}

// NewLimiter allows `perMinute` sustained requests with `burst` headroom.
func NewLimiter(perMinute, burst int, enabled bool) *Limiter {
	l := &Limiter{
		entries: make(map[string]*entry),
		rate:    rate.Limit(float64(perMinute) / 60),
		burst:   burst,
		enabled: enabled,
	}
	if enabled {
		go l.sweep()
	}
	return l
}

func (l *Limiter) allow(key string) bool {
	if !l.enabled {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[key]
	if !ok {
		e = &entry{lim: rate.NewLimiter(l.rate, l.burst)}
		l.entries[key] = e
	}
	e.seen = time.Now()
	return e.lim.Allow()
}

func (l *Limiter) sweep() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		l.mu.Lock()
		for k, e := range l.entries {
			if e.seen.Before(cutoff) {
				delete(l.entries, k)
			}
		}
		l.mu.Unlock()
	}
}

// ByIP limits by client IP. Use for login, refresh, invite completion.
func (l *Limiter) ByIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow("ip:" + ClientIP(r)) {
			w.Header().Set("Retry-After", "60")
			httpx.Fail(w, r, httpx.ErrRateLimited)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ByUser limits by authenticated user id, falling back to IP.
func (l *Limiter) ByUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := "ip:" + ClientIP(r)
		if uid := userIDFromContext(r.Context()); uid != "" {
			key = "user:" + uid
		}
		if !l.allow(key) {
			w.Header().Set("Retry-After", "60")
			httpx.Fail(w, r, httpx.ErrRateLimited)
			return
		}
		next.ServeHTTP(w, r)
	})
}
