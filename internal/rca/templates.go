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

// Agent-specific prompts
const logsAgentPrompt = `You are a log forensics specialist. Analyze the provided container logs for a Kubernetes incident.
Identify: error patterns, crash signatures, resource exhaustion indicators, and timing of failures.
Cite exact log lines as evidence. Return ONLY valid JSON:
{"findings": "what the logs reveal", "severity": "low|medium|high|critical", "evidence": ["exact log line 1", "exact log line 2"], "confidence": 0.0-1.0}`

const eventsAgentPrompt = `You are a Kubernetes events specialist. Analyze the provided cluster events for an incident.
Identify: scheduling issues, probe failures, OOM kills, image pull failures, and node conditions.
Cite exact event messages as evidence. Return ONLY valid JSON:
{"findings": "what the events reveal", "severity": "low|medium|high|critical", "evidence": ["exact event 1", "exact event 2"], "confidence": 0.0-1.0}`

const deployAgentPrompt = `You are a deployment reliability specialist. Analyze this incident for deployment-related causes.
Consider: recent rollouts, configuration changes, resource limit changes, and replica counts.
Based on the incident type and message, determine if a deployment change triggered this. Return ONLY valid JSON:
{"findings": "deployment analysis", "severity": "low|medium|high|critical", "evidence": ["relevant observation 1"], "confidence": 0.0-1.0}`

const synthesisPrompt = `You are a lead incident investigator. Three specialists have independently analyzed the same incident.

Logs specialist findings:
{{ .LogsFindings }}

Events specialist findings:
{{ .EventsFindings }}

Deployment specialist findings:
{{ .DeployFindings }}

Previous similar incidents:
{{ .PastLearnings }}

Synthesize these independent analyses into a single, definitive root cause analysis. Resolve any disagreements between specialists. Return ONLY valid JSON:
{
  "summary": "one-line summary",
  "root_cause": "detailed explanation reconciling all findings",
  "confidence": 0.0-1.0,
  "suggested_actions": "comma-separated: restart,scale_up,scale_down,rollback,delete_pod",
  "evidence_logs": ["combined log evidence"],
  "evidence_events": ["combined event evidence"]
}`

type synthesisData struct {
	LogsFindings   string
	EventsFindings string
	DeployFindings string
	PastLearnings  string
}

func BuildSynthesisPrompt(logsFindings, eventsFindings, deployFindings, pastLearnings string) string {
	tmpl, err := template.New("synthesis").Parse(synthesisPrompt)
	if err != nil {
		return synthesisPrompt
	}
	data := synthesisData{
		LogsFindings:   logsFindings,
		EventsFindings: eventsFindings,
		DeployFindings: deployFindings,
		PastLearnings:  pastLearnings,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return synthesisPrompt
	}
	return buf.String()
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
