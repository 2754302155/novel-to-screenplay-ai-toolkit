package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/repository"
)

var (
	ErrInvalidTaskInput = errors.New("invalid conversion task input")
	ErrTaskNotFound     = repository.ErrTaskNotFound
)

type CreateConversionTaskInput struct {
	SourceText string
	Chapters   []domain.Chapter
}

type TaskService struct {
	repository *repository.TaskRepository
	now        func() time.Time
}

func NewTaskService(repository *repository.TaskRepository) *TaskService {
	return &TaskService{repository: repository, now: time.Now}
}

func (service *TaskService) Create(input CreateConversionTaskInput) (domain.ConversionTask, error) {
	if len(input.Chapters) < 3 {
		return domain.ConversionTask{}, ErrInvalidTaskInput
	}

	now := service.now().UTC()
	task := domain.ConversionTask{
		ID:         newTaskID(),
		Status:     domain.TaskStatusPending,
		Progress:   5,
		Stage:      "任务已创建，等待进入转换流程。",
		SourceText: strings.TrimSpace(input.SourceText),
		Chapters:   input.Chapters,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return service.repository.Save(task), nil
}

func (service *TaskService) Get(id string) (domain.ConversionTask, error) {
	task, err := service.repository.FindByID(id)
	if err != nil {
		return domain.ConversionTask{}, err
	}

	return service.advance(task), nil
}

func (service *TaskService) advance(task domain.ConversionTask) domain.ConversionTask {
	now := service.now().UTC()
	elapsed := now.Sub(task.CreatedAt)

	switch {
	case elapsed >= 4*time.Second:
		task.Status = domain.TaskStatusValidating
		task.Progress = 80
		task.Stage = "正在校验 YAML Schema 和任务结果。"
	case elapsed >= 2*time.Second:
		task.Status = domain.TaskStatusProcessing
		task.Progress = 45
		task.Stage = "正在整理章节内容并准备剧本转换。"
	default:
		task.Status = domain.TaskStatusPending
		task.Progress = 5
		task.Stage = "任务已创建，等待进入转换流程。"
	}

	task.UpdatedAt = now
	return service.repository.Save(task)
}

func newTaskID() string {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "task-" + time.Now().UTC().Format("20060102150405")
	}

	return "task-" + hex.EncodeToString(random)
}
