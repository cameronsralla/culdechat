package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxRequestBytes = 2 << 20 // 2 MiB

// LimitBodySize rejects oversized JSON/multipart bodies.
func LimitBodySize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
		}
		c.Next()
	}
}

// ApplyTrustedProxies uses CULDECHAT_TRUSTED_PROXIES (comma-separated) or trusts none.
func ApplyTrustedProxies(engine *gin.Engine) {
	raw := strings.TrimSpace(os.Getenv("CULDECHAT_TRUSTED_PROXIES"))
	if raw == "" {
		_ = engine.SetTrustedProxies(nil)
		return
	}
	parts := strings.Split(raw, ",")
	proxies := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			proxies = append(proxies, p)
		}
	}
	_ = engine.SetTrustedProxies(proxies)
}
