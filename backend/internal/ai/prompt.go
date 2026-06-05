package ai

import (
	"fmt"
	"strings"
)

func RenderScreenplayPrompt(input DraftInput) string {
	chapterLines := make([]string, 0, len(input.Chapters))
	for _, chapter := range input.Chapters {
		chapterLines = append(chapterLines, fmt.Sprintf("- %s %s，约 %d 字", chapter.ID, chapter.Title, chapter.WordCount))
	}

	return fmt.Sprintf(`请将以下中文小说章节转换为结构化 YAML 剧本初稿。

要求：
- 输出 schema_version、project、source、adaptation、characters、scenes、quality_report。
- 每个场景必须包含 source_refs。
- 每个场景至少包含一个 beats 节拍，节拍 type 只能使用 action、dialogue、voice_over、transition、note。
- scenes[].characters 必须引用 characters[].id。
- confidence 必须是 0 到 1 之间的数字。
- 对推断内容写入 warnings 或 human_review_required。
- 不要生成最终拍摄剧本，只生成可编辑初稿。
- 只输出 YAML，不要输出 Markdown 代码围栏。

章节：
%s

正文：
%s`, strings.Join(chapterLines, "\n"), input.SourceText)
}
