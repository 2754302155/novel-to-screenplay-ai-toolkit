package service

import (
	"context"
	"errors"
	"sync"
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
	if task.Status != domain.TaskStatusValidating {
		t.Fatalf("expected validating status while background job starts, got %s", task.Status)
	}

	task = waitForTaskStatus(t, repo, task.ID, domain.TaskStatusCompleted)
	if task.Draft == nil || task.YAML == "" {
		t.Fatalf("expected generated draft and yaml")
	}
}

func TestTaskServiceStartsCompletionOnce(t *testing.T) {
	repo := repository.NewTaskRepository()
	client := &blockingClient{ready: make(chan struct{})}
	service := NewTaskService(repo, client)
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
	if _, err := service.Get(task.ID); err != nil {
		t.Fatalf("first get: %v", err)
	}
	if _, err := service.Get(task.ID); err != nil {
		t.Fatalf("second get: %v", err)
	}

	client.waitForCall(t)
	if client.callCount() != 1 {
		t.Fatalf("expected one ai call, got %d", client.callCount())
	}
	close(client.ready)
	waitForTaskStatus(t, repo, task.ID, domain.TaskStatusCompleted)
}

func sampleTaskChapters() []domain.Chapter {
	return []domain.Chapter{
		{ID: "CH001", Title: "第一章", WordCount: 10},
		{ID: "CH002", Title: "第二章", WordCount: 10},
		{ID: "CH003", Title: "第三章", WordCount: 10},
	}
}

func waitForTaskStatus(
	t *testing.T,
	repo *repository.TaskRepository,
	taskID string,
	status domain.ConversionTaskStatus,
) domain.ConversionTask {
	t.Helper()

	for range 100 {
		task, err := repo.FindByID(taskID)
		if err != nil {
			t.Fatalf("find task: %v", err)
		}
		if task.Status == status {
			return task
		}
		time.Sleep(10 * time.Millisecond)
	}

	task, _ := repo.FindByID(taskID)
	t.Fatalf("expected status %s, got %s", status, task.Status)
	return domain.ConversionTask{}
}

type blockingClient struct {
	ready chan struct{}
	mu    sync.Mutex
	calls int
}

func (client *blockingClient) GenerateDraft(ctx context.Context, input ai.DraftInput) (domain.ScreenplayDraft, error) {
	client.mu.Lock()
	client.calls++
	client.mu.Unlock()

	select {
	case <-ctx.Done():
		return domain.ScreenplayDraft{}, ctx.Err()
	case <-client.ready:
	}

	return ai.NewLocalClient().GenerateDraft(ctx, input)
}

func (client *blockingClient) callCount() int {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.calls
}

func (client *blockingClient) waitForCall(t *testing.T) {
	t.Helper()

	for range 100 {
		if client.callCount() > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("expected ai call")
}
