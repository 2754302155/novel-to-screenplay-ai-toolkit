package service

import "testing"

func TestChapterParserParsesChineseHeadings(t *testing.T) {
	parser := NewChapterParser()
	result := parser.Parse(ParseChaptersInput{Text: `第一章 旧笔记
林夏翻开旧笔记。

第二章 雨夜
雨越下越大。

第三章 来客
门外有人敲门。`})

	if len(result.BlockingErrors) != 0 {
		t.Fatalf("expected no blocking errors, got %+v", result.BlockingErrors)
	}
	if len(result.Chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(result.Chapters))
	}
	if result.Chapters[0].ID != "CH001" || result.Chapters[0].Title != "第一章 旧笔记" {
		t.Fatalf("unexpected first chapter: %+v", result.Chapters[0])
	}
}

func TestChapterParserBlocksLessThanThreeChapters(t *testing.T) {
	parser := NewChapterParser()
	result := parser.Parse(ParseChaptersInput{Text: `第一章
正文。

第二章
正文。`})

	if len(result.BlockingErrors) == 0 {
		t.Fatal("expected blocking error")
	}
	if result.BlockingErrors[0].Code != "CHAPTER_COUNT_TOO_LOW" {
		t.Fatalf("unexpected error code: %s", result.BlockingErrors[0].Code)
	}
}

func TestChapterParserCleansNoise(t *testing.T) {
	parser := NewChapterParser()
	result := parser.Parse(ParseChaptersInput{Text: `第一章
请收藏本站


正文。`})

	if result.Chapters[0].Body != "正文。" {
		t.Fatalf("unexpected cleaned body: %q", result.Chapters[0].Body)
	}
}
