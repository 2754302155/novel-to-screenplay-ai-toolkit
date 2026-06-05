package schema

import (
	"fmt"
	"strings"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

const lowConfidenceThreshold = 0.5

func GenerateQualityReport(
	draft domain.ScreenplayDraft,
	chapters []domain.Chapter,
	validation ValidationResult,
) domain.QualityReport {
	report := draft.QualityReport
	if report.Warnings == nil {
		report.Warnings = []string{}
	}
	if report.HumanReviewRequired == nil {
		report.HumanReviewRequired = []string{}
	}

	knownChapters := map[string]bool{}
	for _, chapter := range chapters {
		if chapter.ID != "" {
			knownChapters[chapter.ID] = true
		}
	}
	if len(knownChapters) == 0 {
		for _, chapter := range draft.Source.Chapters {
			if chapter.ID != "" {
				knownChapters[chapter.ID] = true
			}
		}
	}

	coveredChapters := map[string]bool{}
	lowConfidenceBeats := 0
	placeholderBeats := 0
	for _, scene := range draft.Scenes {
		for _, sourceRef := range scene.SourceRefs {
			if len(knownChapters) == 0 || knownChapters[sourceRef] {
				coveredChapters[sourceRef] = true
			}
		}
		for _, beat := range scene.Beats {
			if beat.Confidence > 0 && beat.Confidence < lowConfidenceThreshold {
				lowConfidenceBeats++
			}
			if isPlaceholderText(beat.Text) {
				placeholderBeats++
			}
		}
	}

	totalChapters := len(knownChapters)
	if totalChapters == 0 {
		totalChapters = draft.Source.ChapterCount
	}
	report.Coverage.ConvertedChapters = len(coveredChapters)
	if totalChapters > 0 {
		report.Coverage.EstimatedUnconvertedRate = 1 - float64(len(coveredChapters))/float64(totalChapters)
		if report.Coverage.EstimatedUnconvertedRate < 0 {
			report.Coverage.EstimatedUnconvertedRate = 0
		}
	}

	if len(draft.Scenes) == 0 {
		report.Warnings = append(report.Warnings, "未生成任何场景。")
		report.HumanReviewRequired = append(report.HumanReviewRequired, "需要人工补充场景列表。")
	}
	if totalChapters > 0 && len(coveredChapters) < totalChapters {
		report.Warnings = append(report.Warnings, fmt.Sprintf("章节覆盖不足：已覆盖 %d/%d 章。", len(coveredChapters), totalChapters))
		report.HumanReviewRequired = append(report.HumanReviewRequired, "检查未覆盖章节是否需要补生成场景。")
	}
	if lowConfidenceBeats > 0 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("存在 %d 个低置信度节拍。", lowConfidenceBeats))
		report.HumanReviewRequired = append(report.HumanReviewRequired, "复核低置信度动作、对白和旁白。")
	}
	if placeholderBeats > 0 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("存在 %d 个占位节拍文本。", placeholderBeats))
		report.HumanReviewRequired = append(report.HumanReviewRequired, "补充占位节拍的具体动作、对白或转场。")
	}
	if !validation.Valid {
		for _, issue := range validation.Issues {
			report.Warnings = append(report.Warnings, issue.Path+"："+issue.Message)
		}
		report.HumanReviewRequired = append(report.HumanReviewRequired, "修复 YAML Schema 校验问题后再继续编辑。")
	}

	report.Warnings = uniqueStrings(report.Warnings)
	report.HumanReviewRequired = uniqueStrings(report.HumanReviewRequired)
	return report
}

func isPlaceholderText(text string) bool {
	trimmed := strings.TrimSpace(text)
	return trimmed == "" ||
		strings.Contains(trimmed, "待人工补充") ||
		strings.Contains(trimmed, "待补充") ||
		strings.Contains(trimmed, "占位")
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		unique = append(unique, trimmed)
	}
	return unique
}
