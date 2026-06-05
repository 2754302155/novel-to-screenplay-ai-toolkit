package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/ai"
)

type TestAIRequest struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
}

func registerAIRoutes(router *gin.RouterGroup) {
	router.POST("/ai/test", func(ctx *gin.Context) {
		var request TestAIRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"ok":      false,
				"message": "请求格式不正确，请检查 AI 配置。",
			})
			return
		}

		client, err := ai.NewOpenAICompatibleClient(ai.ProviderConfig{
			Provider: request.Provider,
			BaseURL:  request.BaseURL,
			Model:    request.Model,
			APIKey:   request.APIKey,
		})
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"ok":      false,
				"message": "AI 配置不完整，请填写 Base URL、模型名和 API Key。",
			})
			return
		}

		testCtx, cancel := context.WithTimeout(ctx.Request.Context(), 30*time.Second)
		defer cancel()
		if err := client.TestConnection(testCtx); err != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{
				"ok":      false,
				"message": "AI 联通测试失败：" + err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"message": "AI 联通测试成功。",
		})
	})
}
