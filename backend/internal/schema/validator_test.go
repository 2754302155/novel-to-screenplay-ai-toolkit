package schema

import (
	"testing"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

func TestValidateDraftRejectsMissingScenes(t *testing.T) {
	result := ValidateDraft(domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Source:        domain.Source{ChapterCount: 3},
		Characters:    []domain.Character{{ID: "CHAR001"}},
	})

	if result.Valid {
		t.Fatal("expected invalid draft")
	}
}

func TestRepairDraftFillsBasicFields(t *testing.T) {
	repaired := RepairDraft(domain.ScreenplayDraft{}, sampleChapters(), time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC))

	if repaired.SchemaVersion != domain.CurrentSchemaVersion {
		t.Fatalf("unexpected schema version: %s", repaired.SchemaVersion)
	}
	if repaired.Source.ChapterCount != 3 {
		t.Fatalf("expected chapter count 3, got %d", repaired.Source.ChapterCount)
	}
	if result := ValidateDraft(repaired); !result.Valid {
		t.Fatalf("expected repaired draft to be valid, got %#v", result.Issues)
	}
}

func TestRepairDraftFixesInvalidSceneReferences(t *testing.T) {
	repaired := RepairDraft(domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Source:        domain.Source{ChapterCount: 1},
		Characters: []domain.Character{
			{Name: "林夏"},
		},
		Scenes: []domain.Scene{
			{
				Characters: []string{"missing"},
				Beats: []domain.Beat{
					{Type: "unknown", Text: "", Confidence: 2},
				},
			},
		},
	}, sampleChapters(), time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC))

	if result := ValidateDraft(repaired); !result.Valid {
		t.Fatalf("expected repaired draft to be valid, got %#v", result.Issues)
	}
	if repaired.Scenes[0].Beats[0].Type != "note" {
		t.Fatalf("expected invalid beat type to be repaired, got %s", repaired.Scenes[0].Beats[0].Type)
	}
}

func sampleChapters() []domain.Chapter {
	return []domain.Chapter{
		{ID: "CH001", Title: "第一章", WordCount: 10},
		{ID: "CH002", Title: "第二章", WordCount: 10},
		{ID: "CH003", Title: "第三章", WordCount: 10},
	}
}
