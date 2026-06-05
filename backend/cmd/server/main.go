package main

import (
	"log"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/api"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/config"
)

func main() {
	cfg := config.Load()
	router := api.NewRouter(cfg)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
