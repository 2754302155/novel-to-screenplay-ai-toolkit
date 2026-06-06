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

func TestValidateDraftRejectsInvalidRefsAndBeats(t *testing.T) {
	result := ValidateDraft(domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Source: domain.Source{
			ChapterCount: 3,
			Chapters:     sampleChapters(),
		},
		Characters: []domain.Character{{ID: "CHAR001"}},
		Scenes: []domain.Scene{
			{
				ID:         "SCENE001",
				SourceRefs: []string{"CH999"},
				Characters: []string{"MISSING"},
				Beats: []domain.Beat{
					{Type: "bad", Text: "", Confidence: 2},
				},
			},
		},
	})

	if result.Valid {
		t.Fatal("expected invalid draft")
	}
	if len(result.Issues) < 4 {
		t.Fatalf("expected multiple validation issues, got %#v", result.Issues)
	}
}

func TestGenerateQualityReportFlagsCoverageAndLowConfidence(t *testing.T) {
	draft := domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Source: domain.Source{
			ChapterCount: 3,
			Chapters:     sampleChapters(),
		},
		Characters: []domain.Character{{ID: "CHAR001"}},
		Scenes: []domain.Scene{
			{
				ID:         "SCENE001",
				SourceRefs: []string{"CH001"},
				Characters: []string{"CHAR001"},
				Beats: []domain.Beat{
					{Type: "note", Text: "待人工补充。", Confidence: 0.3},
				},
			},
		},
	}

	report := GenerateQualityReport(draft, sampleChapters(), ValidateDraft(draft))

	if report.Coverage.ConvertedChapters != 1 {
		t.Fatalf("expected 1 converted chapter, got %d", report.Coverage.ConvertedChapters)
	}
	if report.Coverage.EstimatedUnconvertedRate <= 0 {
		t.Fatalf("expected unconverted rate, got %f", report.Coverage.EstimatedUnconvertedRate)
	}
	if len(report.Warnings) == 0 || len(report.HumanReviewRequired) == 0 {
		t.Fatalf("expected warnings and human review items, got %#v", report)
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

func TestRepairDraftReplacesUnknownSourceRefs(t *testing.T) {
	repaired := RepairDraft(domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Source: domain.Source{
			ChapterCount: 3,
			Chapters:     sampleChapters(),
		},
		Characters: []domain.Character{{ID: "CHAR001", Name: "林夏"}},
		Scenes: []domain.Scene{
			{
				ID:         "SCENE001",
				SourceRefs: []string{"CH999"},
				Characters: []string{"CHAR001"},
				Beats: []domain.Beat{
					{Type: "action", Text: "林夏进入书店。", Confidence: 0.8},
				},
			},
		},
	}, sampleChapters(), time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC))

	if result := ValidateDraft(repaired); !result.Valid {
		t.Fatalf("expected repaired draft to be valid, got %#v", result.Issues)
	}
	if got := repaired.Scenes[0].SourceRefs; len(got) != 1 || got[0] != "CH001" {
		t.Fatalf("expected unknown source ref to be replaced with CH001, got %#v", got)
	}
}

func sampleChapters() []domain.Chapter {
	return []domain.Chapter{
		{ID: "CH001", Title: "第一章", WordCount: 10},
		{ID: "CH002", Title: "第二章", WordCount: 10},
		{ID: "CH003", Title: "第三章", WordCount: 10},
	}
}
