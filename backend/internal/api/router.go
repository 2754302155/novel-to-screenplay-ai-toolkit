package api

import (
	"net/http"

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
	registerTaskRoutes(api, service.NewTaskService(repository.NewTaskRepository(), ai.NewLocalClient()))

	return router
}
