package ai

import (
	"context"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

type DraftInput struct {
	SourceText string
	Chapters   []domain.Chapter
}

type Client interface {
	GenerateDraft(ctx context.Context, input DraftInput) (domain.ScreenplayDraft, error)
}
