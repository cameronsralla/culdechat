package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/gin-gonic/gin"
)

// NewRouter constructs the gin.Engine with all routes and middleware registered.
func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())

	api := router.Group("/api")

	// Health endpoint under /api
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register sub-route groups under /api
	RegisterAuthRoutes(api)
	RegisterBoardRoutes(api)
	RegisterPostRoutes(api)
	RegisterCommentRoutes(api)
	RegisterReactionRoutes(api)
	RegisterProfileRoutes(api)

	return router
}
