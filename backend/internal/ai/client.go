package ai

import (
	"context"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

type DraftInput struct {
	SourceText string
	Chapters   []domain.Chapter
}

type ProviderConfig struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
}

type Client interface {
	GenerateDraft(ctx context.Context, input DraftInput) (domain.ScreenplayDraft, error)
}
