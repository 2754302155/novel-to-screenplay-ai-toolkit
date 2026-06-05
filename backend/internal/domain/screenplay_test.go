package domain

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestScreenplayDraftSerializesToYAML(t *testing.T) {
	draft := ScreenplayDraft{
		SchemaVersion: CurrentSchemaVersion,
		Project: Project{
			Title:       "未命名作品",
			Author:      "",
			GeneratedAt: time.Date(2026, 6, 5, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
		},
		Source: Source{
			ChapterCount: 3,
			Chapters: []Chapter{
				{ID: "CH001", Title: "第一章", WordCount: 4200},
				{ID: "CH002", Title: "第二章", WordCount: 3900},
				{ID: "CH003", Title: "第三章", WordCount: 4500},
			},
		},
		Adaptation: Adaptation{
			Format:   "web_drama",
			Logline:  "主人公被迫面对旧日秘密。",
			Synopsis: "前三章建立主要冲突。",
			Themes:   []string{"成长", "秘密"},
		},
		Characters: []Character{
			{
				ID:              "CHAR001",
				Name:            "林夏",
				Aliases:         []string{},
				RoleType:        "protagonist",
				Description:     "故事主人公。",
				FirstAppearance: "CH001",
			},
		},
		Scenes: []Scene{
			{
				ID:              "SC001",
				SourceRefs:      []string{"CH001"},
				Heading:         "内景 - 林夏的房间 - 夜",
				Location:        "林夏的房间",
				TimeOfDay:       "night",
				Characters:      []string{"CHAR001"},
				DramaticPurpose: "建立主人公困境",
				Beats: []Beat{
					{Type: "action", Text: "林夏翻开旧笔记。", Confidence: 0.86},
				},
				Notes: []string{"需确认道具后续是否出现。"},
			},
		},
		QualityReport: QualityReport{
			Coverage: Coverage{
				ConvertedChapters:        3,
				EstimatedUnconvertedRate: 0.12,
			},
			Warnings:            []string{"部分心理描写被转换为动作。"},
			HumanReviewRequired: []string{"确认 SC001 中对白是否符合角色语气。"},
		},
	}

	content, err := yaml.Marshal(draft)
	if err != nil {
		t.Fatalf("marshal draft: %v", err)
	}

	var decoded ScreenplayDraft
	if err := yaml.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("unmarshal draft: %v", err)
	}

	if decoded.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("expected schema version %s, got %s", CurrentSchemaVersion, decoded.SchemaVersion)
	}
	if decoded.Source.ChapterCount != 3 {
		t.Fatalf("expected chapter count 3, got %d", decoded.Source.ChapterCount)
	}
}
