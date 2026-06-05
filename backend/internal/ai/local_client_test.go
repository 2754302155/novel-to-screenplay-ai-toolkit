package ai

import (
	"context"
	"testing"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
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
