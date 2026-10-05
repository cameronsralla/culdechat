package main

// @title Cul-de-Chat API
// @version 1.0
// @description Private town-square API for a residential community.
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"log"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/routes"
	"github.com/cameronsralla/culdechat/utils"
)

func main() {
	if _, err := utils.LoadRootDotEnv(); err != nil {
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

	ctx := context.Background()
	if err := utils.RequireJWTSecret(); err != nil {
		log.Fatalf("jwt config: %v", err)
	}
	if _, err := postgres.Initialize(ctx); err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	if err := models.Migrate(ctx); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}
	if err := models.BootstrapAdmin(ctx); err != nil {
		log.Fatalf("bootstrap admin failed: %v", err)
	}

	router := routes.NewRouter()

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
