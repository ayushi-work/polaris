package rca

import (
	"bytes"
	"text/template"

	"github.com/ayushi/polaris/internal/models"
)

const promptTemplate = `Analyze the following Kubernetes incident:

Incident ID: {{ .Incident.ID }}
Type: {{ .Incident.IncidentType }}
Severity: {{ .Incident.Severity }}
Resource: {{ .Incident.ResourceName }}
Namespace: {{ .Incident.Namespace }}
Message: {{ .Incident.Message }}

## Recent Logs
{{ .Logs }}

## Kubernetes Events
{{ .Events }}

Based on the above information, what is the root cause of this incident? What remediation actions do you recommend?`

type promptData struct {
	Incident *models.Incident
	Logs     string
	Events   string
}

func BuildPrompt(incident *models.Incident, logs, events string) string {
	tmpl, err := template.New("rca").Parse(promptTemplate)
	if err != nil {
		return promptTemplate
	}

	data := promptData{
		Incident: incident,
		Logs:     logs,
		Events:   events,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return promptTemplate
	}

	return buf.String()
}
