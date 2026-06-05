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
- beats[].text 必须写具体动作、对白、旁白或转场内容，不能留空，不能写“待补充”。
- aliases、themes、source_refs、characters、notes、warnings、human_review_required 必须输出 YAML 字符串数组；没有内容时输出 []，禁止输出 true/false。
- human_review_required 只能是字符串数组，用于说明需要人工复核的事项。
- 对推断内容写入 warnings 或 human_review_required 字符串数组。
- 不要生成最终拍摄剧本，只生成可编辑初稿。
- 只输出 YAML，不要输出 Markdown 代码围栏。

章节：
%s

正文：
%s`, strings.Join(chapterLines, "\n"), input.SourceText)
}
