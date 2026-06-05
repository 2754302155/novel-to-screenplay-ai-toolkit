package ai

import (
	"context"
	"testing"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

func TestLocalClientGeneratesDraft(t *testing.T) {
	client := NewLocalClient()
	draft, err := client.GenerateDraft(context.Background(), DraftInput{
		SourceText: "第一章林夏翻开旧笔记。",
		Chapters: []domain.Chapter{
			{ID: "CH001", Title: "第一章", WordCount: 10},
			{ID: "CH002", Title: "第二章", WordCount: 10},
			{ID: "CH003", Title: "第三章", WordCount: 10},
		},
	})

	if err != nil {
		t.Fatalf("generate draft: %v", err)
	}
	if draft.SchemaVersion != domain.CurrentSchemaVersion {
		t.Fatalf("unexpected schema version: %s", draft.SchemaVersion)
	}
	if len(draft.Scenes) != 3 {
		t.Fatalf("expected 3 scenes, got %d", len(draft.Scenes))
	}
}
