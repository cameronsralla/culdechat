package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterPostRoutes registers post related endpoints under /posts.
func RegisterPostRoutes(r gin.IRouter) {
	service := &services.PostService{}

	grp := r.Group("/posts")

	grp.GET("/board/:board_id", func(c *gin.Context) {
		boardIDStr := c.Param("board_id")
		boardID, err := uuid.Parse(boardIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board_id"})
			return
		}
		out, err := service.ListByBoard(c.Request.Context(), boardID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	grp.GET("/bulletins", func(c *gin.Context) {
		out, err := service.ListBulletins(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	grp.POST("", middleware.AuthRequired(), func(c *gin.Context) {
		var in services.CreatePostInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userIDStr := c.GetString("user_id")
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		// Simple isAdmin check: reload user and read IsAdmin
		isAdmin := false
		if u, err := models.GetUserByID(c.Request.Context(), userUUID); err == nil && u != nil {
			isAdmin = u.IsAdmin
		}
		out, err := service.Create(c.Request.Context(), userUUID, in, isAdmin)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, out)
	})
}
