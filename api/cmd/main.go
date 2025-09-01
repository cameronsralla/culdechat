package main

import (
	"log"
	"net/http"

	"context"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/routes"
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

	// Initialize Postgres connection pool and ensure core tables
	ctx := context.Background()
	if _, err := postgres.Initialize(ctx); err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	if err := models.EnsureUsersTable(ctx); err != nil {
		log.Fatalf("ensure users table failed: %v", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	// Register routes
	routes.RegisterAuthRoutes(router)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
