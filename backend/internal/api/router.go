package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/ai"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/config"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/repository"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/service"
)

func NewRouter(cfg config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	api := router.Group("/api")
	api.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "novel-to-screenplay-api",
			"version": cfg.Version,
		})
	})
	registerAIRoutes(api)
	registerChapterRoutes(api)
	registerYAMLRoutes(api)
	registerTaskRoutes(api, service.NewTaskService(newTaskRepository(cfg), ai.NewLocalClient()))

	return router
}

func newTaskRepository(cfg config.Config) *repository.TaskRepository {
	if cfg.DatabaseURL == "" {
		return repository.NewTaskRepository()
	}

	var lastErr error
	for attempt := 1; attempt <= 10; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		taskRepository, err := repository.NewPostgresBackedTaskRepository(ctx, cfg.DatabaseURL)
		cancel()
		if err == nil {
			return taskRepository
		}

		lastErr = err
		log.Printf("postgres task repository unavailable, retrying (%d/10): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	panic(fmt.Sprintf("initialize postgres task repository: %v", lastErr))
}
