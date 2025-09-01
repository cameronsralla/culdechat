package middleware

import (
	"net/http"
	"strings"

	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
)

// AuthRequired validates Authorization: Bearer <token> and sets user claims in context.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseAndValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// Stash claims for handlers
		c.Set("user_id", claims.UserID)
		c.Set("unit", claims.Unit)
		c.Next()
	}
}
