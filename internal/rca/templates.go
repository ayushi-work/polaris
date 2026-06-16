package rca

import (
	"bytes"
	"text/template"

	"github.com/ayushi/polaris/internal/models"
)

const promptTemplate = `Analyze the following Kubernetes incident and return a JSON response.

Incident ID: {{ .Incident.ID }}
Type: {{ .Incident.IncidentType }}
Severity: {{ .Incident.Severity }}
Resource: {{ .Incident.ResourceName }}
Namespace: {{ .Incident.Namespace }}
Message: {{ .Incident.Message }}
{{ if .PastLearnings }}
## Previous Similar Incidents (learn from these)
{{ .PastLearnings }}
{{ end }}
## Recent Logs
{{ .Logs }}

## Kubernetes Events
{{ .Events }}

Return ONLY valid JSON (no markdown, no backticks):

{
  "summary": "one-line summary",
  "root_cause": "detailed explanation of what happened and why",
  "confidence": 0.0-1.0,
  "suggested_actions": "comma-separated: restart,scale_up,scale_down,rollback,delete_pod",
  "evidence_logs": ["log line 1 that proves this", "log line 2 that supports this"],
  "evidence_events": ["event 1 that correlates", "event 2 that confirms the timeline"]
}

For evidence_logs and evidence_events: copy the EXACT lines from the logs and events above that support your diagnosis. Do not paraphrase. If no relevant evidence exists, use empty arrays.`

type promptData struct {
	Incident      *models.Incident
	Logs          string
	Events        string
	PastLearnings string
}

func BuildPrompt(incident *models.Incident, logs, events, pastLearnings string) string {
	tmpl, err := template.New("rca").Parse(promptTemplate)
	if err != nil {
		return promptTemplate
	}

	data := promptData{
		Incident:      incident,
		Logs:          logs,
		Events:        events,
		PastLearnings: pastLearnings,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return promptTemplate
	}

	return buf.String()
}
