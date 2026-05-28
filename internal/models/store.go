package models

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewID(prefix string) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(b))
}

type IncidentFilter struct {
	Status   string
	Severity string
	Service  string
	Limit    int
	Offset   int
}

type Store interface {
	CreateIncident(ctx context.Context, inc *Incident) error
	GetIncident(ctx context.Context, id string) (*Incident, error)
	ListIncidents(ctx context.Context, filter IncidentFilter) ([]Incident, error)
	UpdateIncident(ctx context.Context, inc *Incident) error
	DeleteIncident(ctx context.Context, id string) error
	CountIncidents(ctx context.Context, status string) (int64, error)

	CreateRemediation(ctx context.Context, rem *Remediation) error
	GetRemediation(ctx context.Context, id string) (*Remediation, error)
	ListRemediations(ctx context.Context, incidentID string) ([]Remediation, error)
	UpdateRemediation(ctx context.Context, rem *Remediation) error

	CreateRCAResult(ctx context.Context, r *RCAResult) error
	GetRCAResult(ctx context.Context, incidentID string) (*RCAResult, error)

	CreateChaosScenario(ctx context.Context, s *ChaosScenario) error
	GetChaosScenario(ctx context.Context, id string) (*ChaosScenario, error)
	ListChaosScenarios(ctx context.Context) ([]ChaosScenario, error)
	UpdateChaosScenario(ctx context.Context, s *ChaosScenario) error
	DeleteChaosScenario(ctx context.Context, id string) error
}

type SQLiteStore struct {
	db *gorm.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(path+"?_journal_mode=WAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.AutoMigrate(&Incident{}, &Remediation{}, &RCAResult{}, &ChaosScenario{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) CreateIncident(ctx context.Context, inc *Incident) error {
	return s.db.WithContext(ctx).Create(inc).Error
}

func (s *SQLiteStore) GetIncident(ctx context.Context, id string) (*Incident, error) {
	var inc Incident
	err := s.db.WithContext(ctx).
		Preload("Remediations").
		Preload("RCAResult").
		First(&inc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &inc, nil
}

func (s *SQLiteStore) ListIncidents(ctx context.Context, filter IncidentFilter) ([]Incident, error) {
	var incs []Incident
	q := s.db.WithContext(ctx).Order("detected_at DESC")

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if filter.Service != "" {
		q = q.Where("resource_name = ?", filter.Service)
	}
	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	} else {
		q = q.Limit(50)
	}
	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}

	err := q.Find(&incs).Error
	return incs, err
}

func (s *SQLiteStore) UpdateIncident(ctx context.Context, inc *Incident) error {
	return s.db.WithContext(ctx).Save(inc).Error
}

func (s *SQLiteStore) DeleteIncident(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&Incident{}, "id = ?", id).Error
}

func (s *SQLiteStore) CountIncidents(ctx context.Context, status string) (int64, error) {
	var count int64
	q := s.db.WithContext(ctx).Model(&Incident{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Count(&count).Error
	return count, err
}

func (s *SQLiteStore) CreateRemediation(ctx context.Context, rem *Remediation) error {
	return s.db.WithContext(ctx).Create(rem).Error
}

func (s *SQLiteStore) GetRemediation(ctx context.Context, id string) (*Remediation, error) {
	var rem Remediation
	err := s.db.WithContext(ctx).First(&rem, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rem, nil
}

func (s *SQLiteStore) ListRemediations(ctx context.Context, incidentID string) ([]Remediation, error) {
	var rems []Remediation
	q := s.db.WithContext(ctx).Order("created_at DESC")
	if incidentID != "" {
		q = q.Where("incident_id = ?", incidentID)
	}
	err := q.Find(&rems).Error
	return rems, err
}

func (s *SQLiteStore) UpdateRemediation(ctx context.Context, rem *Remediation) error {
	return s.db.WithContext(ctx).Save(rem).Error
}

func (s *SQLiteStore) CreateRCAResult(ctx context.Context, r *RCAResult) error {
	return s.db.WithContext(ctx).Create(r).Error
}

func (s *SQLiteStore) GetRCAResult(ctx context.Context, incidentID string) (*RCAResult, error) {
	var r RCAResult
	err := s.db.WithContext(ctx).First(&r, "incident_id = ?", incidentID).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SQLiteStore) CreateChaosScenario(ctx context.Context, sc *ChaosScenario) error {
	return s.db.WithContext(ctx).Create(sc).Error
}

func (s *SQLiteStore) GetChaosScenario(ctx context.Context, id string) (*ChaosScenario, error) {
	var sc ChaosScenario
	err := s.db.WithContext(ctx).First(&sc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

func (s *SQLiteStore) ListChaosScenarios(ctx context.Context) ([]ChaosScenario, error) {
	var scs []ChaosScenario
	err := s.db.WithContext(ctx).Order("created_at DESC").Find(&scs).Error
	return scs, err
}

func (s *SQLiteStore) UpdateChaosScenario(ctx context.Context, sc *ChaosScenario) error {
	return s.db.WithContext(ctx).Save(sc).Error
}

func (s *SQLiteStore) DeleteChaosScenario(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&ChaosScenario{}, "id = ?", id).Error
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
