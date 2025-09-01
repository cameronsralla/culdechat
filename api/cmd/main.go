package main

import (
	"log"
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load .env from repository root before anything else
	if _, err := utils.LoadRootDotEnv(); err != nil {
		// Non-fatal: continue even if no .env was found
		log.Printf("warning: %v", err)
	}

	_, closer, err := utils.Init()
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	defer func() {
		if closer != nil {
			_ = closer.Close()
		}
	}()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
