package domain

import "time"

type ConversionTaskStatus string

const (
	TaskStatusPending    ConversionTaskStatus = "pending"
	TaskStatusProcessing ConversionTaskStatus = "processing"
	TaskStatusValidating ConversionTaskStatus = "validating"
	TaskStatusCompleted  ConversionTaskStatus = "completed"
	TaskStatusFailed     ConversionTaskStatus = "failed"
)

type ConversionTask struct {
	ID                string               `json:"id"`
	Status            ConversionTaskStatus `json:"status"`
	Progress          int                  `json:"progress"`
	Stage             string               `json:"stage"`
	SourceText        string               `json:"source_text,omitempty"`
	Chapters          []Chapter            `json:"chapters"`
	AIConfig          AIConfig             `json:"-"`
	GenerationStarted bool                 `json:"-"`
	TotalChunks       int                  `json:"total_chunks,omitempty"`
	CompletedChunks   int                  `json:"completed_chunks,omitempty"`
	CurrentChunk      string               `json:"current_chunk,omitempty"`
	Draft             *ScreenplayDraft     `json:"draft,omitempty"`
	YAML              string               `json:"yaml,omitempty"`
	ErrorMessage      string               `json:"error_message,omitempty"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

type AIConfig struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key,omitempty"`
}
