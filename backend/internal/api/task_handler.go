package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/service"
)

type CreateConversionTaskRequest struct {
	SourceText string           `json:"source_text"`
	Chapters   []domain.Chapter `json:"chapters"`
}

func registerTaskRoutes(router *gin.RouterGroup, taskService *service.TaskService) {
	router.POST("/conversion-tasks", func(ctx *gin.Context) {
		var request CreateConversionTaskRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_REQUEST",
				"message": "请求格式不正确，请确认章节后重试。",
			})
			return
		}

		task, err := taskService.Create(service.CreateConversionTaskInput{
			SourceText: request.SourceText,
			Chapters:   request.Chapters,
		})
		if errors.Is(err, service.ErrInvalidTaskInput) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"code":    "CHAPTER_COUNT_TOO_LOW",
				"message": "章节不足 3 章，暂不能创建转换任务。",
			})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    "TASK_CREATE_FAILED",
				"message": "转换任务创建失败，请稍后重试。",
			})
			return
		}

		ctx.JSON(http.StatusCreated, task)
	})

	router.GET("/conversion-tasks/:id", func(ctx *gin.Context) {
		task, err := taskService.Get(ctx.Param("id"))
		if errors.Is(err, service.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    "TASK_NOT_FOUND",
				"message": "未找到该转换任务。",
			})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    "TASK_QUERY_FAILED",
				"message": "转换任务查询失败，请稍后重试。",
			})
			return
		}

		ctx.JSON(http.StatusOK, task)
	})
}
