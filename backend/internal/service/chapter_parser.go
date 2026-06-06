package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const minChapterCount = 3

var chapterHeadingPattern = regexp.MustCompile(`^\s*(第[一二三四五六七八九十百千万零〇两\d]+[章节回集卷][^\n\r]*)\s*$`)

type ParseChaptersInput struct {
	Text string
}

type ParseChaptersResult struct {
	Chapters       []ParsedChapter `json:"chapters"`
	CleanedText    string          `json:"cleaned_text,omitempty"`
	OriginalChars  int             `json:"original_chars"`
	CleanedChars   int             `json:"cleaned_chars"`
	ChineseRatio   float64         `json:"chinese_ratio"`
	Warnings       []ParseIssue    `json:"warnings"`
	BlockingErrors []ParseIssue    `json:"blocking_errors"`
}

type ParsedChapter struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	WordCount     int    `json:"word_count"`
	Body          string `json:"body"`
	InferredTitle bool   `json:"inferred_title"`
}

type ParseIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ChapterParser struct{}

func NewChapterParser() ChapterParser {
	return ChapterParser{}
}

func (parser ChapterParser) Parse(input ParseChaptersInput) ParseChaptersResult {
	originalText := strings.TrimSpace(input.Text)
	cleanedText := cleanNovelText(originalText)
	chapters := splitChapters(cleanedText)

	result := ParseChaptersResult{
		Chapters:      chapters,
		CleanedText:   cleanedText,
		OriginalChars: countTextChars(originalText),
		CleanedChars:  countTextChars(cleanedText),
		ChineseRatio:  chineseRatio(cleanedText),
		Warnings:      []ParseIssue{},
	}

	if len(chapters) < minChapterCount {
		result.BlockingErrors = append(result.BlockingErrors, ParseIssue{
			Code:    "CHAPTER_COUNT_TOO_LOW",
			Message: "章节不足 3 章，请补充至少 3 个章节后再转换。",
		})
	}

	if cleanedText == "" {
		result.BlockingErrors = append(result.BlockingErrors, ParseIssue{
			Code:    "EMPTY_TEXT",
			Message: "未检测到有效小说正文，请粘贴或上传文本后再解析。",
		})
	}

	if cleanedText != "" && result.ChineseRatio < 0.35 {
		result.Warnings = append(result.Warnings, ParseIssue{
			Code:    "LOW_CHINESE_RATIO",
			Message: "文本中中文内容比例较低，AI 剧本转换质量可能下降。",
		})
	}

	return result
}

func cleanNovelText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	lastBlank := false

	for _, line := range lines {
		normalized := strings.TrimSpace(line)
		if shouldDropLine(normalized) {
			continue
		}

		if normalized == "" {
			if lastBlank {
				continue
			}
			lastBlank = true
			cleaned = append(cleaned, "")
			continue
		}

		lastBlank = false
		cleaned = append(cleaned, normalized)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func shouldDropLine(line string) bool {
	if line == "" {
		return false
	}

	lower := strings.ToLower(line)
	dropKeywords := []string{
		"本章未完",
		"请收藏",
		"求推荐",
		"求月票",
		"手机用户请浏览",
		"www.",
		"http://",
		"https://",
	}

	for _, keyword := range dropKeywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

func splitChapters(text string) []ParsedChapter {
	if text == "" {
		return []ParsedChapter{}
	}

	type segment struct {
		title         string
		bodyLines     []string
		inferredTitle bool
	}

	lines := strings.Split(text, "\n")
	segments := []segment{}
	current := segment{title: "未命名章节", inferredTitle: true}
	hasContent := false

	for _, line := range lines {
		title := chapterTitle(line)
		if title != "" {
			if hasContent {
				segments = append(segments, current)
			}
			current = segment{title: title, inferredTitle: false}
			hasContent = true
			continue
		}

		current.bodyLines = append(current.bodyLines, line)
		if strings.TrimSpace(line) != "" {
			hasContent = true
		}
	}

	if hasContent {
		segments = append(segments, current)
	}

	chapters := make([]ParsedChapter, 0, len(segments))
	for index, chapter := range segments {
		title := strings.TrimSpace(chapter.title)
		inferred := chapter.inferredTitle
		if title == "" || title == "未命名章节" {
			title = fmt.Sprintf("第%d章", index+1)
			inferred = true
		}

		body := strings.TrimSpace(strings.Join(chapter.bodyLines, "\n"))
		chapters = append(chapters, ParsedChapter{
			ID:            fmt.Sprintf("CH%03d", index+1),
			Title:         title,
			WordCount:     countTextChars(body),
			Body:          body,
			InferredTitle: inferred,
		})
	}

	return chapters
}

func chapterTitle(line string) string {
	match := chapterHeadingPattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return ""
	}

	return strings.TrimSpace(match[1])
}

func countTextChars(text string) int {
	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			count++
		}
	}

	return count
}

func chineseRatio(text string) float64 {
	total := 0
	chinese := 0

	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		text = text[size:]
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsDigit(r) {
			continue
		}

		total++
		if unicode.Is(unicode.Han, r) {
			chinese++
		}
	}

	if total == 0 {
		return 0
	}

	return float64(chinese) / float64(total)
}
