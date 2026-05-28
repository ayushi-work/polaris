package models

import (
	"time"

	"gorm.io/gorm"
)

type Incident struct {
	ID           string     `gorm:"primaryKey;size:36" json:"id"`
	Cluster      string     `gorm:"size:255" json:"cluster"`
	Namespace    string     `gorm:"size:255;index" json:"namespace"`
	Kind         string     `gorm:"size:64" json:"kind"`
	ResourceName string     `gorm:"size:255" json:"resource_name"`
	IncidentType string     `gorm:"size:64;index" json:"incident_type"`
	Severity     string     `gorm:"size:32;index" json:"severity"`
	Status       string     `gorm:"size:32;index" json:"status"`
	Message      string     `gorm:"type:text" json:"message"`
	Details      string     `gorm:"type:text" json:"details"`
	DetectedAt   time.Time  `gorm:"index" json:"detected_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Remediations []Remediation `gorm:"foreignKey:IncidentID" json:"remediations,omitempty"`
	RCAResult    *RCAResult    `gorm:"foreignKey:IncidentID" json:"rca_result,omitempty"`
}

func (i *Incident) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = NewID("INC")
	}
	if i.Status == "" {
		i.Status = "detected"
	}
	if i.DetectedAt.IsZero() {
		i.DetectedAt = time.Now().UTC()
	}
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *Incident) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now().UTC()
	return nil
}
