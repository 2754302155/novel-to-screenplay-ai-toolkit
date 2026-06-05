package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
	"gopkg.in/yaml.v3"
)

func TestLocalClientGeneratesDraft(t *testing.T) {
	client := NewLocalClient()
	draft, err := client.GenerateDraft(context.Background(), DraftInput{
		SourceText: "第一章林夏翻开旧笔记。",
		Chapters: []domain.Chapter{
			{ID: "CH001", Title: "第一章", WordCount: 10},
			{ID: "CH002", Title: "第二章", WordCount: 10},
			{ID: "CH003", Title: "第三章", WordCount: 10},
		},
	})

	if err != nil {
		t.Fatalf("generate draft: %v", err)
	}
	if draft.SchemaVersion != domain.CurrentSchemaVersion {
		t.Fatalf("unexpected schema version: %s", draft.SchemaVersion)
	}
	if len(draft.Scenes) != 3 {
		t.Fatalf("expected 3 scenes, got %d", len(draft.Scenes))
	}
}

func TestOpenAICompatibleClientRequiresConfig(t *testing.T) {
	_, err := NewOpenAICompatibleClient(ProviderConfig{Model: "gpt-4.1"})
	if err == nil {
		t.Fatal("expected missing api key error")
	}
}

func TestOpenAICompatibleClientNormalizesEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "uses default",
			baseURL:  "",
			expected: defaultOpenAICompatibleBaseURL + "/v1/chat/completions",
		},
		{
			name:     "adds v1 chat completions",
			baseURL:  "https://api.openai.com",
			expected: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "adds chat completions after existing v1",
			baseURL:  "https://api.openai.com/v1",
			expected: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "trims trailing slash after v1",
			baseURL:  "https://api.openai.com/v1/",
			expected: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "keeps custom endpoint after v1",
			baseURL:  "https://api.openai.com/v1/custom/completions",
			expected: "https://api.openai.com/v1/custom/completions",
		},
		{
			name:     "trims custom endpoint trailing slash",
			baseURL:  "https://api.openai.com/v1/custom/completions/",
			expected: "https://api.openai.com/v1/custom/completions",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := NewOpenAICompatibleClient(ProviderConfig{
				BaseURL: test.baseURL,
				Model:   "gpt-4.1",
				APIKey:  "test-key",
			})
			if err != nil {
				t.Fatalf("create client: %v", err)
			}
			if client.config.BaseURL != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, client.config.BaseURL)
			}
		})
	}
}

func TestExtractYAMLContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "plain yaml",
			content:  "schema_version: \"1.0\"\ncharacters: []",
			expected: "schema_version: \"1.0\"\ncharacters: []",
		},
		{
			name:     "yaml fence",
			content:  "```yaml\nschema_version: \"1.0\"\ncharacters: []\n```",
			expected: "schema_version: \"1.0\"\ncharacters: []",
		},
		{
			name:     "uppercase fence with intro",
			content:  "下面是 YAML：\n```YAML\nschema_version: \"1.0\"\ncharacters: []\n```",
			expected: "schema_version: \"1.0\"\ncharacters: []",
		},
		{
			name:     "intro without fence",
			content:  "下面是结果：\nschema_version: \"1.0\"\ncharacters: []",
			expected: "schema_version: \"1.0\"\ncharacters: []",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := extractYAMLContent(test.content); actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestOpenAICompatibleClientSendsMaxTokens(t *testing.T) {
	var requestBody chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"schema_version: \"1.0\"\nproject:\n  title: 测试\n  author: 未知\n  generated_at: \"2026-06-05T10:00:00Z\"\nsource:\n  chapter_count: 3\n  chapters: []\nadaptation:\n  format: 短剧\n  logline: 测试\n  synopsis: 测试\n  themes: []\ncharacters: []\nscenes: []\nquality_report:\n  coverage:\n    converted_chapters: 0\n    estimated_unconverted_ratio: 0\n  warnings: []\n  human_review_required: []"}}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClient(ProviderConfig{
		BaseURL: server.URL + "/v1",
		Model:   "demo",
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	_, err = client.GenerateDraft(context.Background(), DraftInput{
		SourceText: "正文",
		Chapters:   []domain.Chapter{{ID: "CH001", Title: "第一章", WordCount: 10}},
	})
	if err != nil {
		t.Fatalf("generate draft: %v", err)
	}
	if requestBody.MaxTokens != 4096 {
		t.Fatalf("expected max tokens 4096, got %d", requestBody.MaxTokens)
	}
}

func TestNormalizeDraftYAMLCoercesStringLists(t *testing.T) {
	content := `schema_version: "1.0"
characters:
  - id: CHAR001
    name: 林夏
    aliases: 林
    role_type: protagonist
    description: 主角
    first_appearance: CH001
scenes:
  - id: SCENE001
    source_refs: CH001
    characters: CHAR001
    beats: []
quality_report:
  warnings: "需要检查"
  human_review_required: true
`

	normalized, err := normalizeDraftYAML(content)
	if err != nil {
		t.Fatalf("normalize yaml: %v", err)
	}

	var draft domain.ScreenplayDraft
	if err := yaml.Unmarshal(normalized, &draft); err != nil {
		t.Fatalf("unmarshal normalized yaml: %v", err)
	}
	if len(draft.Characters) != 1 || draft.Characters[0].ID != "CHAR001" {
		t.Fatalf("expected root characters to stay as character objects, got %#v", draft.Characters)
	}
	if len(draft.Characters[0].Aliases) != 1 || draft.Characters[0].Aliases[0] != "林" {
		t.Fatalf("expected aliases to be coerced, got %#v", draft.Characters[0].Aliases)
	}
	if len(draft.Scenes[0].SourceRefs) != 1 || draft.Scenes[0].SourceRefs[0] != "CH001" {
		t.Fatalf("expected source refs to be coerced, got %#v", draft.Scenes[0].SourceRefs)
	}
	if len(draft.QualityReport.HumanReviewRequired) != 1 {
		t.Fatalf("expected human review flag to be coerced, got %#v", draft.QualityReport.HumanReviewRequired)
	}
}
