package models

import (
	"time"

	"gorm.io/gorm"
)

type Remediation struct {
	ID           string     `gorm:"primaryKey;size:36" json:"id"`
	IncidentID   string     `gorm:"size:36;index" json:"incident_id"`
	Type         string     `gorm:"size:64" json:"type"`
	Status       string     `gorm:"size:32" json:"status"`
	TargetKind   string     `gorm:"size:64" json:"target_kind"`
	TargetName   string     `gorm:"size:255" json:"target_name"`
	Namespace    string     `gorm:"size:255" json:"namespace"`
	Parameters   string     `gorm:"type:text" json:"parameters"`
	Output       string     `gorm:"type:text" json:"output"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	IsAutomated  bool       `json:"is_automated"`
	CreatedAt    time.Time  `json:"created_at"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

func (r *Remediation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = NewID("REM")
	}
	r.CreatedAt = time.Now().UTC()
	return nil
}
