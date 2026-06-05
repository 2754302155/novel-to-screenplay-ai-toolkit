package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/service"
)

type CreateConversionTaskRequest struct {
	SourceText string               `json:"source_text"`
	Chapters   []TaskChapterRequest `json:"chapters"`
	AIConfig   domain.AIConfig      `json:"ai_config"`
}

type TaskChapterRequest struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	WordCount int    `json:"word_count"`
	Body      string `json:"body"`
}

type ConversionTaskSummary struct {
	ID              string                      `json:"id"`
	Status          domain.ConversionTaskStatus `json:"status"`
	Progress        int                         `json:"progress"`
	Stage           string                      `json:"stage"`
	ChapterCount    int                         `json:"chapter_count"`
	TotalChunks     int                         `json:"total_chunks,omitempty"`
	CompletedChunks int                         `json:"completed_chunks,omitempty"`
	CurrentChunk    string                      `json:"current_chunk,omitempty"`
	ErrorMessage    string                      `json:"error_message,omitempty"`
	CreatedAt       string                      `json:"created_at"`
	UpdatedAt       string                      `json:"updated_at"`
}

func registerTaskRoutes(router *gin.RouterGroup, taskService *service.TaskService) {
	router.GET("/conversion-tasks", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"tasks": summarizeTasks(taskService.List()),
		})
	})

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
			Chapters:   request.toDomainChapters(),
			AIConfig:   request.AIConfig,
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

func (request CreateConversionTaskRequest) toDomainChapters() []domain.Chapter {
	chapters := make([]domain.Chapter, 0, len(request.Chapters))
	for _, chapter := range request.Chapters {
		chapters = append(chapters, domain.Chapter{
			ID:        chapter.ID,
			Title:     chapter.Title,
			WordCount: chapter.WordCount,
			Body:      chapter.Body,
		})
	}
	return chapters
}

func summarizeTasks(tasks []domain.ConversionTask) []ConversionTaskSummary {
	summaries := make([]ConversionTaskSummary, 0, len(tasks))
	for _, task := range tasks {
		summaries = append(summaries, ConversionTaskSummary{
			ID:              task.ID,
			Status:          task.Status,
			Progress:        task.Progress,
			Stage:           task.Stage,
			ChapterCount:    len(task.Chapters),
			TotalChunks:     task.TotalChunks,
			CompletedChunks: task.CompletedChunks,
			CurrentChunk:    task.CurrentChunk,
			ErrorMessage:    task.ErrorMessage,
			CreatedAt:       task.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:       task.UpdatedAt.Format(time.RFC3339Nano),
		})
	}
	return summaries
}
