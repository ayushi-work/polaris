package models

import (
	"time"

	"gorm.io/gorm"
)

type ChaosScenario struct {
	ID          string     `gorm:"primaryKey;size:36" json:"id"`
	Name        string     `gorm:"size:255" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Action      string     `gorm:"size:64" json:"action"`
	Target      string     `gorm:"type:text" json:"target"`
	Parameters  string     `gorm:"type:text" json:"parameters"`
	Schedule    string     `gorm:"size:128" json:"schedule"`
	Enabled     bool       `json:"enabled"`
	RunCount    int        `json:"run_count"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (s *ChaosScenario) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = NewID("CHAOS")
	}
	s.CreatedAt = time.Now().UTC()
	return nil
}
