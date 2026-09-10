package middleware

import (
	"net/http"
	"strings"

	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthRequired validates Authorization: Bearer <token>, loads the user, and requires an active account.
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
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		u, err := models.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}
		if u == nil || u.Status != models.UserStatusActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account is not active"})
			return
		}
		c.Set("user_id", u.ID.String())
		c.Set("unit", u.UnitNumber)
		c.Set("is_admin", u.IsAdmin)
		c.Next()
	}
}

// AdminRequired requires a valid JWT whose subject is a currently active business admin.
func AdminRequired() gin.HandlerFunc {
	auth := AuthRequired()
	return func(c *gin.Context) {
		auth(c)
		if c.IsAborted() {
			return
		}
		isAdmin, _ := c.Get("is_admin")
		ok, _ := isAdmin.(bool)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
			return
		}
		c.Next()
	}
}
