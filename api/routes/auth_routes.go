package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterAuthRoutes registers authentication-related routes under /auth.
func RegisterAuthRoutes(r gin.IRouter) {
	svc := &services.AuthService{}

	auth := r.Group("/auth")

	auth.POST("/register", func(c *gin.Context) {
		var in services.RegisterInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		out, err := svc.Register(c.Request.Context(), in)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, out)
	})

	auth.POST("/login", func(c *gin.Context) {
		var in services.LoginInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		out, err := svc.Login(c.Request.Context(), in)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	auth.GET("/me", middleware.AuthRequired(), func(c *gin.Context) {
		userIDStr := c.GetString("user_id")
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}

		u, err := models.GetUserByID(c.Request.Context(), userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}
		if u == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          u.ID.String(),
			"email":       u.Email,
			"unit_number": u.UnitNumber,
			"status":      u.Status,
		})
	})
}
