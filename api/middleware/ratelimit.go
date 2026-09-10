package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
)

type limiterBucket struct {
	times []time.Time
}

type memoryLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*limiterBucket
	n        int
	window   time.Duration
	lastSweep time.Time
}

func newMemoryLimiter(n int, window time.Duration) *memoryLimiter {
	return &memoryLimiter{
		buckets: map[string]*limiterBucket{},
		n:       n,
		window:  window,
	}
}

func (l *memoryLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) > l.window {
		for k, b := range l.buckets {
			cut := now.Add(-l.window)
			kept := b.times[:0]
			for _, t := range b.times {
				if t.After(cut) {
					kept = append(kept, t)
				}
			}
			if len(kept) == 0 {
				delete(l.buckets, k)
			} else {
				b.times = kept
			}
		}
		l.lastSweep = now
	}

	b := l.buckets[key]
	if b == nil {
		b = &limiterBucket{}
		l.buckets[key] = b
	}
	cut := now.Add(-l.window)
	kept := b.times[:0]
	for _, t := range b.times {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	b.times = kept
	if len(b.times) >= l.n {
		return false
	}
	b.times = append(b.times, now)
	return true
}

func rateLimitDisabled() bool {
	return strings.EqualFold(os.Getenv("CULDECHAT_RATE_LIMIT"), "off")
}

// RateLimit limits requests per client IP for a route. Disabled when CULDECHAT_RATE_LIMIT=off.
func RateLimit(max int, window time.Duration) gin.HandlerFunc {
	lim := newMemoryLimiter(max, window)
	return func(c *gin.Context) {
		if rateLimitDisabled() {
			c.Next()
			return
		}
		ip := utils.NormalizeToIPv4(c.ClientIP())
		key := c.FullPath() + "|" + ip
		if !lim.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, try again later"})
			return
		}
		c.Next()
	}
}
