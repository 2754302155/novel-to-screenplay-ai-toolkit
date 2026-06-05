package schema

import (
	"fmt"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

type ValidationIssue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Issues []ValidationIssue `json:"issues"`
}

func ValidateDraft(draft domain.ScreenplayDraft) ValidationResult {
	result := ValidationResult{Valid: true, Issues: []ValidationIssue{}}

	addIssue := func(path string, message string) {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{Path: path, Message: message})
	}

	if draft.SchemaVersion == "" {
		addIssue("schema_version", "Schema 版本不能为空。")
	}
	if draft.Source.ChapterCount < 3 {
		addIssue("source.chapter_count", "输入章节数量必须大于等于 3。")
	}
	if len(draft.Characters) == 0 {
		addIssue("characters", "人物表不能为空。")
	}
	if len(draft.Scenes) == 0 {
		addIssue("scenes", "场景列表不能为空。")
	}

	characters := map[string]bool{}
	for _, character := range draft.Characters {
		if character.ID == "" {
			addIssue("characters[].id", "人物 ID 不能为空。")
			continue
		}
		characters[character.ID] = true
	}

	for index, scene := range draft.Scenes {
		path := fmt.Sprintf("scenes[%d]", index)
		if scene.ID == "" {
			addIssue(path+".id", "场景 ID 不能为空。")
		}
		if len(scene.SourceRefs) == 0 {
			addIssue(path+".source_refs", "每个场景至少需要一个来源章节引用。")
		}
		if len(scene.Beats) == 0 {
			addIssue(path+".beats", "每个场景至少需要一个节拍。")
		}
		for _, characterID := range scene.Characters {
			if !characters[characterID] {
				addIssue(path+".characters", "场景引用了不存在的人物 ID。")
			}
		}
	}

	return result
}

func RepairDraft(draft domain.ScreenplayDraft, chapters []domain.Chapter, now time.Time) domain.ScreenplayDraft {
	if draft.SchemaVersion == "" {
		draft.SchemaVersion = domain.CurrentSchemaVersion
	}
	if draft.Project.Title == "" {
		draft.Project.Title = "未命名作品"
	}
	if draft.Project.GeneratedAt.IsZero() {
		draft.Project.GeneratedAt = now.UTC()
	}
	if draft.Source.ChapterCount == 0 {
		draft.Source.ChapterCount = len(chapters)
	}
	if len(draft.Source.Chapters) == 0 {
		draft.Source.Chapters = chapters
	}
	if len(draft.QualityReport.Warnings) == 0 {
		draft.QualityReport.Warnings = []string{"系统已对 AI 输出执行基础结构修复。"}
	}
	if draft.QualityReport.Coverage.ConvertedChapters == 0 {
		draft.QualityReport.Coverage.ConvertedChapters = len(chapters)
	}

	return draft
}
