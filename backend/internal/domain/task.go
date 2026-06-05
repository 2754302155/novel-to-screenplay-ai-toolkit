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
	ID           string               `json:"id"`
	Status       ConversionTaskStatus `json:"status"`
	Progress     int                  `json:"progress"`
	Stage        string               `json:"stage"`
	SourceText   string               `json:"source_text,omitempty"`
	Chapters     []Chapter            `json:"chapters"`
	Draft        *ScreenplayDraft     `json:"draft,omitempty"`
	ErrorMessage string               `json:"error_message,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}
