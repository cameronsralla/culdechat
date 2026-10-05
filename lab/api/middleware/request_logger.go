package middleware

import (
	"time"

	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
)

// RequestLogger logs basic request/response details for every HTTP request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := utils.NormalizeToIPv4(c.ClientIP())
		userAgent := c.Request.UserAgent()

		// Process request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		utils.Infof("%s %s | status=%d ip=%s ua=%q latency=%s", method, path, status, clientIP, userAgent, latency)
	}
}
