package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/service"
)

type ParseChaptersRequest struct {
	Text string `json:"text"`
}

func registerChapterRoutes(router *gin.RouterGroup) {
	parser := service.NewChapterParser()

	router.POST("/chapters/parse", func(ctx *gin.Context) {
		var request ParseChaptersRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_REQUEST",
				"message": "请求格式不正确，请提交小说正文后重试。",
			})
			return
		}

		result := parser.Parse(service.ParseChaptersInput{Text: request.Text})
		status := http.StatusOK
		if len(result.BlockingErrors) > 0 {
			status = http.StatusUnprocessableEntity
		}

		ctx.JSON(status, result)
	})
}
