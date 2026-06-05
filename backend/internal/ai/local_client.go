package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

type LocalClient struct {
	now func() time.Time
}

func NewLocalClient() *LocalClient {
	return &LocalClient{now: time.Now}
}

func (client *LocalClient) GenerateDraft(ctx context.Context, input DraftInput) (domain.ScreenplayDraft, error) {
	select {
	case <-ctx.Done():
		return domain.ScreenplayDraft{}, ctx.Err()
	default:
	}

	character := domain.Character{
		ID:              "CHAR001",
		Name:            "待确认角色",
		Aliases:         []string{},
		RoleType:        "unknown",
		Description:     "由系统根据输入章节建立的占位人物，需作者在后续编辑中确认姓名、身份和动机。",
		FirstAppearance: firstChapterID(input.Chapters),
	}

	scenes := make([]domain.Scene, 0, len(input.Chapters))
	for index, chapter := range input.Chapters {
		scenes = append(scenes, domain.Scene{
			ID:              sceneID(index + 1),
			SourceRefs:      []string{chapter.ID},
			Heading:         "内景 - 待确认地点 - 时间待确认",
			Location:        "待确认地点",
			TimeOfDay:       "unknown",
			Characters:      []string{character.ID},
			DramaticPurpose: "将章节内容整理为可继续改写的场景初稿。",
			Beats: []domain.Beat{
				{
					Type:       "action",
					Speaker:    "",
					Text:       chapterActionText(input.SourceText, chapter.Title),
					Confidence: 0.62,
				},
				{
					Type:       "note",
					Speaker:    "",
					Text:       "该场景由本地 AI 适配器生成，需在接入真实模型后进一步细化动作、对白和人物动机。",
					Confidence: 0.55,
				},
			},
			Notes: []string{"需确认地点、时间、出场人物和对白归属。"},
		})
	}

	return domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Project: domain.Project{
			Title:       "未命名作品",
			Author:      "",
			GeneratedAt: client.now().UTC(),
		},
		Source: domain.Source{
			ChapterCount: len(input.Chapters),
			Chapters:     input.Chapters,
		},
		Adaptation: domain.Adaptation{
			Format:   "web_drama",
			Logline:  "输入章节已被整理为可继续改写的剧本初稿。",
			Synopsis: "系统根据输入章节建立人物、场景、节拍和质量提示，供作者继续编辑。",
			Themes:   []string{"待确认"},
		},
		Characters: []domain.Character{character},
		Scenes:     scenes,
		Continuity: &domain.Continuity{
			Timeline:          []string{"时间线需在后续章节解析后继续完善。"},
			Foreshadowing:     []string{},
			UnresolvedIssues:  []string{"人物姓名、地点、对白归属和戏剧冲突需人工确认。"},
			CarryForwardNotes: []string{"后续接入真实 AI provider 后应替换占位节拍。"},
		},
		QualityReport: domain.QualityReport{
			Coverage: domain.Coverage{
				ConvertedChapters:        len(input.Chapters),
				EstimatedUnconvertedRate: 0,
			},
			Warnings: []string{
				"当前使用本地 AI 适配器生成结构化初稿，内容偏保守。",
				"部分场景地点、时间和人物信息为待确认项。",
			},
			HumanReviewRequired: []string{
				"确认人物表是否符合原文。",
				"补充每个场景的具体地点、时间和对白。",
			},
		},
	}, nil
}

func firstChapterID(chapters []domain.Chapter) string {
	if len(chapters) == 0 {
		return ""
	}

	return chapters[0].ID
}

func sceneID(number int) string {
	return fmt.Sprintf("SC%03d", number)
}

func chapterActionText(text string, fallback string) string {
	cleaned := strings.TrimSpace(strings.ReplaceAll(text, "\n", ""))
	if cleaned == "" {
		return "根据章节《" + fallback + "》整理动作节拍。"
	}

	delimiters := []string{"。", "！", "？"}
	limit := len(cleaned)
	for _, delimiter := range delimiters {
		if index := strings.Index(cleaned, delimiter); index >= 0 && index+len(delimiter) < limit {
			limit = index + len(delimiter)
		}
	}

	runes := []rune(cleaned)
	if limit > len(cleaned) {
		limit = len(cleaned)
	}

	prefix := []rune(cleaned[:limit])
	if len(prefix) > 80 {
		prefix = runes[:80]
	}

	return "围绕《" + fallback + "》整理关键行动：" + string(prefix)
}
