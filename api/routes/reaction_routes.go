package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterReactionRoutes registers reaction endpoints under /reactions.
func RegisterReactionRoutes(r gin.IRouter) {
	service := &services.ReactionService{}

	grp := r.Group("/reactions")

	grp.POST("", middleware.AuthRequired(), func(c *gin.Context) {
		var in services.ReactInput
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
		if err := service.Upsert(c.Request.Context(), userUUID, in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	grp.DELETE("/:post_id", middleware.AuthRequired(), func(c *gin.Context) {
		postID := c.Param("post_id")
		userIDStr := c.GetString("user_id")
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		if err := service.Remove(c.Request.Context(), userUUID, postID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
