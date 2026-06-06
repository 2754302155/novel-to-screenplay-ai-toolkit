package service

import (
	"context"
	"errors"
	"strings"
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

func TestTaskServiceCompletesWithFallbackChunks(t *testing.T) {
	repo := repository.NewTaskRepository()
	service := NewTaskService(repo, failingClient{})
	startedAt := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return startedAt }

	task, err := service.Create(CreateConversionTaskInput{
		SourceText: "正文",
		Chapters: []domain.Chapter{
			{ID: "CH001", Title: "第一章", WordCount: 10, Body: "第一章正文"},
			{ID: "CH002", Title: "第二章", WordCount: 10, Body: "第二章正文"},
			{ID: "CH003", Title: "第三章", WordCount: 10, Body: "第三章正文"},
		},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	service.now = func() time.Time { return startedAt.Add(7 * time.Second) }
	if _, err := service.Get(task.ID); err != nil {
		t.Fatalf("get task: %v", err)
	}

	task = waitForTaskStatus(t, repo, task.ID, domain.TaskStatusCompleted)
	if task.TotalChunks != 3 || task.CompletedChunks != 3 {
		t.Fatalf("expected 3 completed chunks, got %d/%d", task.CompletedChunks, task.TotalChunks)
	}
	if task.Draft == nil || len(task.Draft.QualityReport.HumanReviewRequired) == 0 {
		t.Fatalf("expected fallback draft with human review warnings")
	}
}

func TestTaskServiceNormalizesChunkSourceRefs(t *testing.T) {
	repo := repository.NewTaskRepository()
	service := NewTaskService(repo, wrongSourceRefClient{})
	startedAt := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return startedAt }

	task, err := service.Create(CreateConversionTaskInput{
		SourceText: "正文",
		Chapters: []domain.Chapter{
			{ID: "CH008", Title: "第七章", WordCount: 10, Body: "第七章正文"},
			{ID: "CH009", Title: "第八章", WordCount: 10, Body: "第八章正文"},
			{ID: "CH010", Title: "第九章", WordCount: 10, Body: "第九章正文"},
			{ID: "CH011", Title: "第十章", WordCount: 10, Body: "第十章正文"},
		},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	service.now = func() time.Time { return startedAt.Add(7 * time.Second) }
	if _, err := service.Get(task.ID); err != nil {
		t.Fatalf("get task: %v", err)
	}

	task = waitForTaskStatus(t, repo, task.ID, domain.TaskStatusCompleted)
	if task.Draft == nil {
		t.Fatal("expected completed draft")
	}
	want := []string{"CH008", "CH009", "CH010", "CH011"}
	if len(task.Draft.Scenes) != len(want) {
		t.Fatalf("expected %d scenes, got %d", len(want), len(task.Draft.Scenes))
	}
	for index, scene := range task.Draft.Scenes {
		if len(scene.SourceRefs) != 1 || scene.SourceRefs[0] != want[index] {
			t.Fatalf("scene %d source refs = %#v, want %s", index, scene.SourceRefs, want[index])
		}
	}
}

func TestBuildDraftChunksSplitsLongChapterBody(t *testing.T) {
	chunks := buildDraftChunks("", []domain.Chapter{
		{ID: "CH001", Title: "第一章", WordCount: 7000, Body: strings.Repeat("甲", maxDraftChunkChars+100)},
		{ID: "CH002", Title: "第二章", WordCount: 10, Body: "乙"},
		{ID: "CH003", Title: "第三章", WordCount: 10, Body: "丙"},
	})

	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}
	if chunks[0].Label != "第一章 分段 1" || chunks[1].Label != "第一章 分段 2" {
		t.Fatalf("expected split labels, got %q and %q", chunks[0].Label, chunks[1].Label)
	}
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

type failingClient struct{}
type wrongSourceRefClient struct{}

func (client failingClient) GenerateDraft(ctx context.Context, input ai.DraftInput) (domain.ScreenplayDraft, error) {
	return domain.ScreenplayDraft{}, errors.New("forced chunk failure")
}

func (client wrongSourceRefClient) GenerateDraft(ctx context.Context, input ai.DraftInput) (domain.ScreenplayDraft, error) {
	return domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Project: domain.Project{
			Title:       "测试",
			GeneratedAt: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC),
		},
		Source: domain.Source{
			ChapterCount: 1,
			Chapters:     []domain.Chapter{{ID: "CH001", Title: "错误章节编号", WordCount: 10}},
		},
		Characters: []domain.Character{
			{ID: "CHAR001", Name: "许七安", Aliases: []string{}, RoleType: "protagonist"},
		},
		Scenes: []domain.Scene{
			{
				ID:         "SCENE001",
				SourceRefs: []string{"CH001"},
				Characters: []string{"CHAR001"},
				Beats: []domain.Beat{
					{Type: "action", Text: "许七安推进剧情。", Confidence: 0.8},
				},
			},
		},
		QualityReport: domain.QualityReport{
			Warnings:            []string{},
			HumanReviewRequired: []string{},
		},
	}, nil
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
