package service

import (
	"errors"
	"testing"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/ai"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/repository"
)

func TestTaskServiceCreatesTask(t *testing.T) {
	service := NewTaskService(repository.NewTaskRepository(), ai.NewLocalClient())
	task, err := service.Create(CreateConversionTaskInput{
		SourceText: "正文",
		Chapters:   sampleTaskChapters(),
	})

	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.ID == "" {
		t.Fatal("expected task id")
	}
	if task.Status != domain.TaskStatusPending {
		t.Fatalf("expected pending status, got %s", task.Status)
	}
}

func TestTaskServiceRejectsInsufficientChapters(t *testing.T) {
	service := NewTaskService(repository.NewTaskRepository(), ai.NewLocalClient())
	_, err := service.Create(CreateConversionTaskInput{
		SourceText: "正文",
		Chapters:   sampleTaskChapters()[:2],
	})

	if !errors.Is(err, ErrInvalidTaskInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestTaskServiceAdvancesStatusByElapsedTime(t *testing.T) {
	repo := repository.NewTaskRepository()
	service := NewTaskService(repo, ai.NewLocalClient())
	startedAt := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return startedAt }

	task, err := service.Create(CreateConversionTaskInput{
		SourceText: "正文",
		Chapters:   sampleTaskChapters(),
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	service.now = func() time.Time { return startedAt.Add(3 * time.Second) }
	task, err = service.Get(task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}

	if task.Status != domain.TaskStatusProcessing {
		t.Fatalf("expected processing status, got %s", task.Status)
	}
}

func TestTaskServiceCompletesWithYAMLDraft(t *testing.T) {
	repo := repository.NewTaskRepository()
	service := NewTaskService(repo, ai.NewLocalClient())
	startedAt := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return startedAt }

	task, err := service.Create(CreateConversionTaskInput{
		SourceText: "第一章林夏翻开旧笔记。第二章雨越下越大。第三章门外有人敲门。",
		Chapters:   sampleTaskChapters(),
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	service.now = func() time.Time { return startedAt.Add(7 * time.Second) }
	task, err = service.Get(task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}

	if task.Status != domain.TaskStatusCompleted {
		t.Fatalf("expected completed status, got %s", task.Status)
	}
	if task.Draft == nil || task.YAML == "" {
		t.Fatalf("expected generated draft and yaml")
	}
}

func sampleTaskChapters() []domain.Chapter {
	return []domain.Chapter{
		{ID: "CH001", Title: "第一章", WordCount: 10},
		{ID: "CH002", Title: "第二章", WordCount: 10},
		{ID: "CH003", Title: "第三章", WordCount: 10},
	}
}
