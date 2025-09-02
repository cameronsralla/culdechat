package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterProfileRoutes registers profile and directory endpoints under /profile and /directory.
func RegisterProfileRoutes(r gin.IRouter) {
	service := &services.ProfileService{}

	profile := r.Group("/profile")

	profile.GET("/me", middleware.AuthRequired(), func(c *gin.Context) {
		userIDStr := c.GetString("user_id")
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		out, err := service.Get(c.Request.Context(), userUUID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	profile.PATCH("/me", middleware.AuthRequired(), func(c *gin.Context) {
		userIDStr := c.GetString("user_id")
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		var in services.UpdateProfileInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		out, err := service.Update(c.Request.Context(), userUUID, in)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	r.GET("/directory", func(c *gin.Context) {
		out, err := service.ListDirectory(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})
}
