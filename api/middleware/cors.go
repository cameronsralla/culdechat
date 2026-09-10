package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var defaultCORSOrigins = []string{
	"http://localhost:8081",
	"http://127.0.0.1:8081",
	"http://localhost:19006",
	"http://127.0.0.1:19006",
}

// CORS allows the Expo web app (and any extra CULDECHAT_CORS_ORIGINS) to call the API.
func CORS() gin.HandlerFunc {
	allowed := allowedOrigins()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && originAllowed(origin, allowed) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "600")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func allowedOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CULDECHAT_CORS_ORIGINS"))
	if raw == "" {
		return defaultCORSOrigins
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts)+len(defaultCORSOrigins))
	out = append(out, defaultCORSOrigins...)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.TrimRight(p, "/"))
		}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	// Expo web may pick a free port during local development.
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
		return true
	}
	return false
}
