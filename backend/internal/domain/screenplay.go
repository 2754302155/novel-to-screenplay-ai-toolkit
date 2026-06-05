package domain

import "time"

const CurrentSchemaVersion = "1.0"

type ScreenplayDraft struct {
	SchemaVersion string        `json:"schema_version" yaml:"schema_version"`
	Project       Project       `json:"project" yaml:"project"`
	Source        Source        `json:"source" yaml:"source"`
	Adaptation    Adaptation    `json:"adaptation" yaml:"adaptation"`
	Characters    []Character   `json:"characters" yaml:"characters"`
	Scenes        []Scene       `json:"scenes" yaml:"scenes"`
	Continuity    *Continuity   `json:"continuity,omitempty" yaml:"continuity,omitempty"`
	QualityReport QualityReport `json:"quality_report" yaml:"quality_report"`
}

type Project struct {
	Title       string    `json:"title" yaml:"title"`
	Author      string    `json:"author" yaml:"author"`
	GeneratedAt time.Time `json:"generated_at" yaml:"generated_at"`
}

type Source struct {
	ChapterCount int       `json:"chapter_count" yaml:"chapter_count"`
	Chapters     []Chapter `json:"chapters" yaml:"chapters"`
}

type Chapter struct {
	ID        string `json:"id" yaml:"id"`
	Title     string `json:"title" yaml:"title"`
	WordCount int    `json:"word_count" yaml:"word_count"`
	Body      string `json:"-" yaml:"-"`
}

type Adaptation struct {
	Format   string   `json:"format" yaml:"format"`
	Logline  string   `json:"logline" yaml:"logline"`
	Synopsis string   `json:"synopsis" yaml:"synopsis"`
	Themes   []string `json:"themes" yaml:"themes"`
}

type Character struct {
	ID              string   `json:"id" yaml:"id"`
	Name            string   `json:"name" yaml:"name"`
	Aliases         []string `json:"aliases" yaml:"aliases"`
	RoleType        string   `json:"role_type" yaml:"role_type"`
	Description     string   `json:"description" yaml:"description"`
	FirstAppearance string   `json:"first_appearance" yaml:"first_appearance"`
}

type Scene struct {
	ID              string   `json:"id" yaml:"id"`
	SourceRefs      []string `json:"source_refs" yaml:"source_refs"`
	Heading         string   `json:"heading" yaml:"heading"`
	Location        string   `json:"location" yaml:"location"`
	TimeOfDay       string   `json:"time_of_day" yaml:"time_of_day"`
	Characters      []string `json:"characters" yaml:"characters"`
	DramaticPurpose string   `json:"dramatic_purpose" yaml:"dramatic_purpose"`
	Beats           []Beat   `json:"beats" yaml:"beats"`
	Notes           []string `json:"notes" yaml:"notes"`
}

type Beat struct {
	Type       string  `json:"type" yaml:"type"`
	Speaker    string  `json:"speaker" yaml:"speaker"`
	Text       string  `json:"text" yaml:"text"`
	Confidence float64 `json:"confidence" yaml:"confidence"`
}

type Continuity struct {
	Timeline          []string `json:"timeline" yaml:"timeline"`
	Foreshadowing     []string `json:"foreshadowing" yaml:"foreshadowing"`
	UnresolvedIssues  []string `json:"unresolved_issues" yaml:"unresolved_issues"`
	CarryForwardNotes []string `json:"carry_forward_notes" yaml:"carry_forward_notes"`
}

type QualityReport struct {
	Coverage            Coverage `json:"coverage" yaml:"coverage"`
	Warnings            []string `json:"warnings" yaml:"warnings"`
	HumanReviewRequired []string `json:"human_review_required" yaml:"human_review_required"`
}

type Coverage struct {
	ConvertedChapters        int     `json:"converted_chapters" yaml:"converted_chapters"`
	EstimatedUnconvertedRate float64 `json:"estimated_unconverted_ratio" yaml:"estimated_unconverted_ratio"`
}
