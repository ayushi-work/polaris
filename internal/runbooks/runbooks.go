package runbooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ayushi/polaris/pkg/iforge"
)

// Runbook is a declarative escalation ladder for an incident type: an ordered
// list of remediation steps. The first step runs immediately (pending →
// approved by human or auto_approve); subsequent steps only auto-create when
// the previous step FAILED, so healing escalates instead of repeating.
type Runbook struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Match       Match  `yaml:"match" json:"match"`
	Steps       []Step `yaml:"steps" json:"steps"`
}

type Match struct {
	IncidentTypes []string `yaml:"incident_types" json:"incident_types"`
	Severities    []string `yaml:"severities" json:"severities"`
}

type Step struct {
	Action      string `yaml:"action" json:"action"` // restart | scale_up | scale_down | rollback | delete_pod
	Params      string `yaml:"params,omitempty" json:"params,omitempty"`
	MinReplicas int    `yaml:"min_replicas,omitempty" json:"min_replicas,omitempty"`
	Wait        string `yaml:"wait,omitempty" json:"wait,omitempty"` // observation wait before next step escalates
}

// Registry loads runbooks from a directory of YAML files.
type Registry struct {
	books []*Runbook
}

func LoadDir(dir string) (*Registry, error) {
	reg := &Registry{}
	if dir == "" {
		return reg, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml")) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read runbook %s: %w", entry.Name(), err)
		}
		var rb Runbook
		if err := yaml.Unmarshal(data, &rb); err != nil {
			return nil, fmt.Errorf("parse runbook %s: %w", entry.Name(), err)
		}
		if rb.Name == "" {
			rb.Name = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		}
		if len(rb.Steps) == 0 {
			continue
		}
		reg.books = append(reg.books, &rb)
	}
	return reg, nil
}

func (r *Registry) Len() int { return len(r.books) }

// ByName looks up a runbook by its name.
func (r *Registry) ByName(name string) *Runbook {
	for _, rb := range r.books {
		if rb.Name == name {
			return rb
		}
	}
	return nil
}

// NextStep returns the escalation step after a failed one, or false.
func (r *Registry) NextStep(rb *Runbook, index int) (Step, bool) {
	if index+1 < len(rb.Steps) {
		return rb.Steps[index+1], true
	}
	return Step{}, false
}

// Select returns the best-matching runbook for the incident, or nil.
func (r *Registry) Select(incidentType, severity string) *Runbook {
	var best *Runbook
	for _, rb := range r.books {
		if !matches(rb, incidentType, severity) {
			continue
		}
		// Prefer runbooks with a severity match and more specific type list.
		if best == nil || score(rb, severity) > score(best, severity) {
			best = rb
		}
	}
	return best
}

func matches(rb *Runbook, incidentType, severity string) bool {
	typeOK := len(rb.Match.IncidentTypes) == 0
	for _, t := range rb.Match.IncidentTypes {
		if strings.EqualFold(t, incidentType) || t == "*" {
			typeOK = true
			break
		}
	}
	sevOK := len(rb.Match.Severities) == 0
	for _, s := range rb.Match.Severities {
		if strings.EqualFold(s, severity) || s == "*" {
			sevOK = true
			break
		}
	}
	return typeOK && sevOK
}

func score(rb *Runbook, severity string) int {
	s := 0
	for _, t := range rb.Match.IncidentTypes {
		if strings.EqualFold(t, iforge.IncidentTypeOOMKilled) || strings.EqualFold(t, "any") {
			s++
		}
	}
	for _, sev := range rb.Match.Severities {
		if strings.EqualFold(sev, severity) {
			s += 2
		}
	}
	return s
}

// FirstStep returns the first remediation of a runbook.
func (rb *Runbook) FirstStep() Step {
	return rb.Steps[0]
}

// NextStep returns the escalation step after a failed one, or nil.
func (rb *Runbook) NextStep(index int) (Step, bool) {
	if index+1 < len(rb.Steps) {
		return rb.Steps[index+1], true
	}
	return Step{}, false
}
