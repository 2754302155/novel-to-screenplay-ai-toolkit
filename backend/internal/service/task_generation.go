package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/ai"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

const (
	maxDraftChunkChars = 6000
	chunkRetryCount    = 2
)

type draftChunk struct {
	Index   int
	Total   int
	Chapter domain.Chapter
	Text    string
	Label   string
}

func (service *TaskService) generateDraft(task domain.ConversionTask, client ai.Client) (domain.ScreenplayDraft, error) {
	chunks := buildDraftChunks(task.SourceText, task.Chapters)
	if len(chunks) == 0 {
		return domain.ScreenplayDraft{}, ErrInvalidTaskInput
	}

	task.TotalChunks = len(chunks)
	task.CompletedChunks = 0
	task.CurrentChunk = chunks[0].Label
	task.Progress = 90
	task.Stage = fmt.Sprintf("准备处理 %d 个文本块。", len(chunks))
	task.UpdatedAt = service.now().UTC()
	task = service.repository.Save(task)

	partials := make([]domain.ScreenplayDraft, 0, len(chunks))
	for index, chunk := range chunks {
		task = service.updateChunkProgress(task, chunk, index)
		partial, err := generateChunkDraft(client, chunk)
		if err != nil {
			partial = fallbackChunkDraft(chunk, err, service.now().UTC())
		}
		partials = append(partials, partial)
	}

	task.CompletedChunks = len(chunks)
	task.CurrentChunk = ""
	task.Progress = 99
	task.Stage = "正在合并分块剧本 YAML。"
	task.UpdatedAt = service.now().UTC()
	service.repository.Save(task)

	return mergeChunkDrafts(task, chunks, partials, service.now().UTC()), nil
}

func (service *TaskService) updateChunkProgress(task domain.ConversionTask, chunk draftChunk, completed int) domain.ConversionTask {
	task.CompletedChunks = completed
	task.CurrentChunk = chunk.Label
	if chunk.Total > 0 {
		task.Progress = 90 + (completed * 9 / chunk.Total)
	}
	task.Stage = fmt.Sprintf("正在处理第 %d/%d 个文本块：%s。", chunk.Index+1, chunk.Total, chunk.Label)
	task.UpdatedAt = service.now().UTC()
	return service.repository.Save(task)
}

func generateChunkDraft(client ai.Client, chunk draftChunk) (domain.ScreenplayDraft, error) {
	var lastErr error
	for attempt := 0; attempt <= chunkRetryCount; attempt++ {
		draft, err := client.GenerateDraft(context.Background(), ai.DraftInput{
			SourceText: chunk.Text,
			Chapters:   []domain.Chapter{withoutChapterBody(chunk.Chapter)},
		})
		if err == nil {
			return draft, nil
		}
		lastErr = err
	}
	return domain.ScreenplayDraft{}, lastErr
}

func buildDraftChunks(sourceText string, chapters []domain.Chapter) []draftChunk {
	chunks := []draftChunk{}
	for _, chapter := range chapters {
		chapterBody := strings.TrimSpace(chapter.Body)
		if chapterBody == "" {
			continue
		}
		for _, text := range splitTextByParagraphs(chapterBody, maxDraftChunkChars) {
			chunks = append(chunks, draftChunk{
				Index:   len(chunks),
				Chapter: chapter,
				Text:    text,
				Label:   chapter.Title,
			})
		}
	}

	if len(chunks) == 0 {
		fallbackChapter := domain.Chapter{ID: "CH001", Title: "全文", WordCount: len([]rune(sourceText))}
		if len(chapters) > 0 {
			fallbackChapter = chapters[0]
		}
		for _, text := range splitTextByParagraphs(sourceText, maxDraftChunkChars) {
			chunks = append(chunks, draftChunk{
				Index:   len(chunks),
				Chapter: fallbackChapter,
				Text:    text,
				Label:   fallbackChapter.Title,
			})
		}
	}

	total := len(chunks)
	labelCounts := map[string]int{}
	for _, chunk := range chunks {
		labelCounts[chunk.Label]++
	}
	labelSeen := map[string]int{}
	for index := range chunks {
		chunks[index].Index = index
		chunks[index].Total = total
		if chunks[index].Label == "" {
			chunks[index].Label = chunks[index].Chapter.ID
		}
		if labelCounts[chunks[index].Label] > 1 {
			labelSeen[chunks[index].Label]++
			chunks[index].Label = fmt.Sprintf("%s 分段 %d", chunks[index].Label, labelSeen[chunks[index].Label])
		}
	}

	return chunks
}

func splitTextByParagraphs(text string, maxChars int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return []string{}
	}
	if len([]rune(trimmed)) <= maxChars {
		return []string{trimmed}
	}

	paragraphs := strings.Split(trimmed, "\n")
	chunks := []string{}
	current := strings.Builder{}
	currentChars := 0

	flush := func() {
		value := strings.TrimSpace(current.String())
		if value != "" {
			chunks = append(chunks, value)
		}
		current.Reset()
		currentChars = 0
	}

	for _, paragraph := range paragraphs {
		normalized := strings.TrimSpace(paragraph)
		if normalized == "" {
			continue
		}
		paragraphChars := len([]rune(normalized))
		if paragraphChars > maxChars {
			flush()
			chunks = append(chunks, splitLongText(normalized, maxChars)...)
			continue
		}
		if currentChars > 0 && currentChars+paragraphChars+1 > maxChars {
			flush()
		}
		if current.Len() > 0 {
			current.WriteString("\n")
			currentChars++
		}
		current.WriteString(normalized)
		currentChars += paragraphChars
	}
	flush()

	return chunks
}

func splitLongText(text string, maxChars int) []string {
	runes := []rune(text)
	chunks := []string{}
	for start := 0; start < len(runes); start += maxChars {
		end := start + maxChars
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}

func mergeChunkDrafts(
	task domain.ConversionTask,
	chunks []draftChunk,
	partials []domain.ScreenplayDraft,
	now time.Time,
) domain.ScreenplayDraft {
	merged := domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Project: domain.Project{
			Title:       "未命名作品",
			GeneratedAt: now.UTC(),
		},
		Source: domain.Source{
			ChapterCount: len(task.Chapters),
			Chapters:     withoutChapterBodies(task.Chapters),
		},
		QualityReport: domain.QualityReport{
			Warnings:            []string{},
			HumanReviewRequired: []string{},
		},
	}

	characterByKey := map[string]string{}
	characterIDByPartial := map[string]string{}
	usedCharacterIDs := map[string]bool{}
	sceneIndex := 1

	for partialIndex, partial := range partials {
		if merged.Project.Title == "未命名作品" && partial.Project.Title != "" {
			merged.Project.Title = partial.Project.Title
		}
		if merged.Project.Author == "" && partial.Project.Author != "" {
			merged.Project.Author = partial.Project.Author
		}
		if merged.Adaptation.Format == "" {
			merged.Adaptation = partial.Adaptation
		}

		for _, character := range partial.Characters {
			originalID := character.ID
			key := characterKey(character)
			canonicalID, exists := characterByKey[key]
			if !exists {
				canonicalID = uniqueCharacterID(character.ID, len(merged.Characters)+1, usedCharacterIDs)
				character.ID = canonicalID
				merged.Characters = append(merged.Characters, character)
				characterByKey[key] = canonicalID
				usedCharacterIDs[canonicalID] = true
			}
			if originalID != "" {
				characterIDByPartial[partialCharacterKey(partialIndex, originalID)] = canonicalID
			}
		}

		for _, scene := range partial.Scenes {
			scene.ID = fmt.Sprintf("SCENE%03d", sceneIndex)
			sceneIndex++
			if len(scene.SourceRefs) == 0 && partialIndex < len(chunks) {
				scene.SourceRefs = []string{chunks[partialIndex].Chapter.ID}
			}
			for index, characterID := range scene.Characters {
				if canonicalID := characterIDByPartial[partialCharacterKey(partialIndex, characterID)]; canonicalID != "" {
					scene.Characters[index] = canonicalID
				}
			}
			merged.Scenes = append(merged.Scenes, scene)
		}

		merged.QualityReport.Warnings = append(merged.QualityReport.Warnings, partial.QualityReport.Warnings...)
		merged.QualityReport.HumanReviewRequired = append(
			merged.QualityReport.HumanReviewRequired,
			partial.QualityReport.HumanReviewRequired...,
		)
	}

	merged.QualityReport.Coverage.ConvertedChapters = len(task.Chapters)
	merged.QualityReport.Coverage.EstimatedUnconvertedRate = 0
	if len(merged.QualityReport.Warnings) == 0 {
		merged.QualityReport.Warnings = []string{"系统已按文本块生成并合并剧本 YAML 初稿。"}
	}

	return merged
}

func fallbackChunkDraft(chunk draftChunk, err error, now time.Time) domain.ScreenplayDraft {
	message := fmt.Sprintf("%s 生成失败，系统已生成占位场景：%v", chunk.Label, err)
	return domain.ScreenplayDraft{
		SchemaVersion: domain.CurrentSchemaVersion,
		Project: domain.Project{
			Title:       "未命名作品",
			GeneratedAt: now.UTC(),
		},
		Source: domain.Source{
			ChapterCount: 1,
			Chapters:     []domain.Chapter{withoutChapterBody(chunk.Chapter)},
		},
		Characters: []domain.Character{
			{
				ID:              "CHAR001",
				Name:            "待确认人物",
				Aliases:         []string{},
				RoleType:        "unknown",
				Description:     "该文本块 AI 生成失败，系统创建占位人物用于保持结构完整。",
				FirstAppearance: chunk.Chapter.ID,
			},
		},
		Scenes: []domain.Scene{
			{
				ID:              "SCENE001",
				SourceRefs:      []string{chunk.Chapter.ID},
				Heading:         chunk.Chapter.Title,
				Location:        "待确认地点",
				TimeOfDay:       "待确认时间",
				Characters:      []string{"CHAR001"},
				DramaticPurpose: "保留失败文本块，等待人工补充。",
				Beats: []domain.Beat{
					{
						Type:       "note",
						Text:       message,
						Confidence: 0.2,
					},
				},
				Notes: []string{"该文本块需要人工复核。"},
			},
		},
		QualityReport: domain.QualityReport{
			Warnings:            []string{message},
			HumanReviewRequired: []string{chunk.Label + " 需要人工补充。"},
		},
	}
}

func withoutChapterBodies(chapters []domain.Chapter) []domain.Chapter {
	sanitized := make([]domain.Chapter, 0, len(chapters))
	for _, chapter := range chapters {
		sanitized = append(sanitized, withoutChapterBody(chapter))
	}
	return sanitized
}

func withoutChapterBody(chapter domain.Chapter) domain.Chapter {
	chapter.Body = ""
	return chapter
}

func characterKey(character domain.Character) string {
	if strings.TrimSpace(character.Name) != "" {
		return strings.TrimSpace(character.Name)
	}
	if strings.TrimSpace(character.ID) != "" {
		return strings.TrimSpace(character.ID)
	}
	return "unknown"
}

func partialCharacterKey(partialIndex int, characterID string) string {
	return fmt.Sprintf("%d:%s", partialIndex, characterID)
}

func uniqueCharacterID(preferred string, index int, used map[string]bool) string {
	candidate := strings.TrimSpace(preferred)
	if candidate == "" {
		candidate = fmt.Sprintf("CHAR%03d", index)
	}
	if !used[candidate] {
		return candidate
	}
	for {
		candidate = fmt.Sprintf("CHAR%03d", index)
		if !used[candidate] {
			return candidate
		}
		index++
	}
}
