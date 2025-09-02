package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterCommentRoutes registers comment related endpoints under /comments.
func RegisterCommentRoutes(r gin.IRouter) {
	service := &services.CommentService{}

	grp := r.Group("/comments")

	grp.GET("/post/:post_id", func(c *gin.Context) {
		postIDStr := c.Param("post_id")
		postID, err := uuid.Parse(postIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
			return
		}
		out, err := service.ListByPost(c.Request.Context(), postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	grp.POST("", middleware.AuthRequired(), func(c *gin.Context) {
		var in services.CreateCommentInput
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
		out, err := service.Create(c.Request.Context(), userUUID, in)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, out)
	})
}
