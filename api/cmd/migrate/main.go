package main

import (
	"context"
	"log"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/models"
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
	if _, err := postgres.Initialize(ctx); err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}

	if err := models.EnsureUsersTable(ctx); err != nil {
		log.Fatalf("ensure users table failed: %v", err)
	}

	utils.Infof("database migrations completed successfully")
}
