package middleware

import (
	"net/http"
	"strings"

	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// authenticate loads an active user from the Bearer token and sets context keys.
// It does not call c.Next(); Gin continues the chain after this middleware returns.
func authenticate(c *gin.Context) bool {
	header := c.GetHeader("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
		return false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	claims, err := utils.ParseAndValidateToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return false
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return false
	}
	u, err := models.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return false
	}
	if u == nil || u.Status != models.UserStatusActive {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account is not active"})
		return false
	}
	c.Set("user_id", u.ID.String())
	c.Set("unit", u.UnitNumber)
	c.Set("is_admin", u.IsAdmin)
	return true
}

// AuthRequired validates Authorization: Bearer <token>, loads the user, and requires an active account.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = authenticate(c)
	}
}

// AdminRequired requires a valid JWT whose subject is a currently active business admin.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticate(c) {
			return
		}
		isAdmin, _ := c.Get("is_admin")
		admin, _ := isAdmin.(bool)
		if !admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
			return
		}
	}
}
