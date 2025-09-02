package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
)

// RegisterBoardRoutes registers board related endpoints under /boards.
func RegisterBoardRoutes(r gin.IRouter) {
	service := &services.BoardService{}

	grp := r.Group("/boards")

	grp.GET("", func(c *gin.Context) {
		out, err := service.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	grp.POST("", middleware.AuthRequired(), func(c *gin.Context) {
		var in services.CreateBoardInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		out, err := service.Create(c.Request.Context(), in)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, out)
	})
}
