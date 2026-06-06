package schema

import (
	"fmt"
	"slices"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

var allowedBeatTypes = []string{"action", "dialogue", "voice_over", "transition", "note"}

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
	if len(draft.Source.Chapters) == 0 {
		addIssue("source.chapters", "来源章节列表不能为空。")
	}
	if len(draft.Characters) == 0 {
		addIssue("characters", "人物表不能为空。")
	}
	if len(draft.Scenes) == 0 {
		addIssue("scenes", "场景列表不能为空。")
	}

	characters := map[string]bool{}
	chapterIDs := map[string]bool{}
	for _, chapter := range draft.Source.Chapters {
		if chapter.ID == "" {
			addIssue("source.chapters[].id", "来源章节 ID 不能为空。")
			continue
		}
		chapterIDs[chapter.ID] = true
	}

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
		for _, sourceRef := range scene.SourceRefs {
			if len(chapterIDs) > 0 && !chapterIDs[sourceRef] {
				addIssue(path+".source_refs", "场景引用了不存在的来源章节 ID。")
			}
		}
		if len(scene.Beats) == 0 {
			addIssue(path+".beats", "每个场景至少需要一个节拍。")
		}
		for _, characterID := range scene.Characters {
			if !characters[characterID] {
				addIssue(path+".characters", "场景引用了不存在的人物 ID。")
			}
		}
		for beatIndex, beat := range scene.Beats {
			beatPath := fmt.Sprintf("%s.beats[%d]", path, beatIndex)
			if !slices.Contains(allowedBeatTypes, beat.Type) {
				addIssue(beatPath+".type", "节拍类型必须是 action、dialogue、voice_over、transition 或 note。")
			}
			if beat.Text == "" {
				addIssue(beatPath+".text", "节拍文本不能为空。")
			}
			if beat.Confidence < 0 || beat.Confidence > 1 {
				addIssue(beatPath+".confidence", "置信度必须在 0 到 1 之间。")
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
	if draft.Source.ChapterCount < len(chapters) {
		draft.Source.ChapterCount = len(chapters)
	}
	if len(draft.Source.Chapters) < len(chapters) {
		draft.Source.Chapters = chapters
	}
	draft.Characters = repairCharacters(draft.Characters)
	draft.Scenes = repairScenes(draft.Scenes, draft.Characters, chapters)
	if len(draft.QualityReport.Warnings) == 0 {
		draft.QualityReport.Warnings = []string{"系统已对 AI 输出执行基础结构修复。"}
	}
	if draft.QualityReport.Coverage.ConvertedChapters == 0 {
		draft.QualityReport.Coverage.ConvertedChapters = len(chapters)
	}

	return draft
}

func repairCharacters(characters []domain.Character) []domain.Character {
	if len(characters) == 0 {
		return []domain.Character{
			{
				ID:              "CHAR001",
				Name:            "待确认人物",
				Aliases:         []string{},
				RoleType:        "unknown",
				Description:     "AI 输出未提供人物表，系统已创建占位人物用于保持剧本结构完整。",
				FirstAppearance: "",
			},
		}
	}

	used := map[string]bool{}
	for index := range characters {
		if characters[index].ID == "" || used[characters[index].ID] {
			characters[index].ID = fmt.Sprintf("CHAR%03d", index+1)
		}
		used[characters[index].ID] = true
		if characters[index].Name == "" {
			characters[index].Name = fmt.Sprintf("待确认人物%d", index+1)
		}
		if characters[index].Aliases == nil {
			characters[index].Aliases = []string{}
		}
		if characters[index].RoleType == "" {
			characters[index].RoleType = "unknown"
		}
	}

	return characters
}

func repairScenes(scenes []domain.Scene, characters []domain.Character, chapters []domain.Chapter) []domain.Scene {
	if len(scenes) == 0 {
		scenes = make([]domain.Scene, 0, len(chapters))
		for index, chapter := range chapters {
			scenes = append(scenes, domain.Scene{
				ID:              fmt.Sprintf("SCENE%03d", index+1),
				SourceRefs:      []string{chapter.ID},
				Heading:         chapter.Title,
				Location:        "待确认地点",
				TimeOfDay:       "待确认时间",
				Characters:      []string{characters[0].ID},
				DramaticPurpose: "根据章节内容生成可编辑剧本初稿。",
				Beats: []domain.Beat{
					fallbackBeat("AI 输出未提供有效场景，系统已创建占位节拍，需人工补充。"),
				},
				Notes: []string{"系统根据章节自动补齐场景结构。"},
			})
		}
		return scenes
	}

	characterIDs := map[string]bool{}
	for _, character := range characters {
		characterIDs[character.ID] = true
	}
	fallbackCharacterID := characters[0].ID
	chapterIDs := map[string]bool{}
	for _, chapter := range chapters {
		chapterIDs[chapter.ID] = true
	}

	for index := range scenes {
		if scenes[index].ID == "" {
			scenes[index].ID = fmt.Sprintf("SCENE%03d", index+1)
		}
		scenes[index].SourceRefs = filterKnownChapterIDs(scenes[index].SourceRefs, chapterIDs)
		if len(scenes[index].SourceRefs) == 0 {
			scenes[index].SourceRefs = []string{chapterIDAt(chapters, index)}
		}
		scenes[index].Characters = filterKnownCharacterIDs(scenes[index].Characters, characterIDs)
		if len(scenes[index].Characters) == 0 {
			scenes[index].Characters = []string{fallbackCharacterID}
		}
		if len(scenes[index].Beats) == 0 {
			scenes[index].Beats = []domain.Beat{fallbackBeat("AI 输出未提供该场景节拍，系统已创建占位节拍，需人工补充。")}
		}
		for beatIndex := range scenes[index].Beats {
			if scenes[index].Beats[beatIndex].Type == "" || !slices.Contains(allowedBeatTypes, scenes[index].Beats[beatIndex].Type) {
				scenes[index].Beats[beatIndex].Type = "note"
			}
			if scenes[index].Beats[beatIndex].Text == "" {
				scenes[index].Beats[beatIndex].Text = "待人工补充。"
			}
			if scenes[index].Beats[beatIndex].Confidence < 0 || scenes[index].Beats[beatIndex].Confidence > 1 {
				scenes[index].Beats[beatIndex].Confidence = 0.3
			}
		}
		if scenes[index].Notes == nil {
			scenes[index].Notes = []string{}
		}
	}

	return scenes
}

func fallbackBeat(text string) domain.Beat {
	return domain.Beat{
		Type:       "note",
		Speaker:    "",
		Text:       text,
		Confidence: 0.3,
	}
}

func chapterIDAt(chapters []domain.Chapter, index int) string {
	if len(chapters) == 0 {
		return "CH001"
	}
	if index >= len(chapters) {
		return chapters[len(chapters)-1].ID
	}
	return chapters[index].ID
}

func filterKnownCharacterIDs(ids []string, known map[string]bool) []string {
	filtered := []string{}
	for _, id := range ids {
		if known[id] {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

func filterKnownChapterIDs(ids []string, known map[string]bool) []string {
	filtered := []string{}
	seen := map[string]bool{}
	for _, id := range ids {
		if known[id] && !seen[id] {
			filtered = append(filtered, id)
			seen[id] = true
		}
	}
	return filtered
}
