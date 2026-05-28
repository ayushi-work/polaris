package models

import (
	"time"

	"gorm.io/gorm"
)

type RCAResult struct {
	ID              string    `gorm:"primaryKey;size:36" json:"id"`
	IncidentID      string    `gorm:"size:36;uniqueIndex" json:"incident_id"`
	Summary         string    `gorm:"type:text" json:"summary"`
	RootCause       string    `gorm:"type:text" json:"root_cause"`
	Confidence      float64   `json:"confidence"`
	LogsSnippet     string    `gorm:"type:text" json:"logs_snippet"`
	EventsSummary   string    `gorm:"type:text" json:"events_summary"`
	SuggestedActions string   `gorm:"type:text" json:"suggested_actions"`
	RawLLMOutput    string    `gorm:"type:text" json:"raw_llm_output"`
	LLMModel        string    `gorm:"size:128" json:"llm_model"`
	PromptTokens    int       `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	CreatedAt       time.Time `json:"created_at"`
}

func (r *RCAResult) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = NewID("RCA")
	}
	r.CreatedAt = time.Now().UTC()
	return nil
}
