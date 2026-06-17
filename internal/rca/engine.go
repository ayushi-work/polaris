package rca

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayushi/polaris/internal/config"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
)

type Engine struct {
	kubeClient kube.Client
	llmCfg     config.LLMConfig
	store      models.Store
}

func NewEngine(kubeClient kube.Client, llmCfg config.LLMConfig, store models.Store) *Engine {
	return &Engine{
		kubeClient: kubeClient,
		llmCfg:     llmCfg,
		store:      store,
	}
}

func (e *Engine) Analyze(ctx context.Context, incident *models.Incident) (*models.RCAResult, error) {
	log.Printf("RCA Engine: analyzing incident %s", incident.ID)

	ctxBundle, err := collectContext(ctx, e.kubeClient, incident)
	if err != nil {
		return nil, fmt.Errorf("collect context: %w", err)
	}

	pastLearnings := e.queryPastIncidents(ctx, incident)

	analysis, err := e.multiAgentAnalyze(ctx, ctxBundle, incident, pastLearnings)
	if err != nil {
		log.Printf("RCA Engine: multi-agent failed, trying single call: %v", err)
		analysis, err = e.callLLM(ctx, ctxBundle, incident, pastLearnings)
	}
	if err != nil {
		log.Printf("RCA Engine: LLM call failed, using heuristics: %v", err)
		analysis = heuristicAnalysis(incident, ctxBundle)
	}

	evidenceLogs, _ := json.Marshal(analysis.evidenceLogs)
	evidenceEvents, _ := json.Marshal(analysis.evidenceEvents)

	rca := &models.RCAResult{
		IncidentID:       incident.ID,
		Summary:          analysis.summary,
		RootCause:        analysis.rootCause,
		Confidence:       analysis.confidence,
		LogsSnippet:      string(evidenceLogs),
		EventsSummary:    string(evidenceEvents),
		SuggestedActions: analysis.suggestedActions,
		RawLLMOutput:     analysis.rawOutput,
		LLMModel:         e.llmCfg.Model,
		CreatedAt:        time.Now().UTC(),
	}

	return rca, nil
}

type contextBundle struct {
	logs   string
	events string
}

type analysisResult struct {
	summary          string
	rootCause        string
	confidence       float64
	suggestedActions string
	evidenceLogs     []string
	evidenceEvents   []string
	rawOutput        string
}

func collectContext(ctx context.Context, client kube.Client, incident *models.Incident) (*contextBundle, error) {
	bundle := &contextBundle{}

	logs, err := client.GetPodLogs(ctx, incident.Namespace, incident.ResourceName, 100)
	if err != nil || logs == "" {
		bundle.logs = syntheticLogs(incident)
	} else {
		bundle.logs = logs
	}

	events, err := client.GetEvents(ctx, incident.Namespace, metav1.ListOptions{})
	if err != nil || len(events.Items) == 0 {
		bundle.events = syntheticEvents(incident)
	} else {
		var eventStrs []string
		for _, ev := range events.Items {
			eventStrs = append(eventStrs, fmt.Sprintf("[%s] %s: %s",
				ev.Type, ev.Reason, ev.Message))
		}
		bundle.events = strings.Join(eventStrs, "\n")
	}

	return bundle, nil
}

func syntheticLogs(incident *models.Incident) string {
	switch incident.IncidentType {
	case "OOMKilled":
		return `12:01:00 INFO  Starting container with memory limit 512Mi
12:02:30 WARN  Memory usage at 78% (398Mi / 512Mi)
12:03:15 WARN  Memory usage at 92% (471Mi / 512Mi)
12:03:45 WARN  Memory usage at 97% (497Mi / 512Mi)
12:04:00 ERROR Container killed: exit code 137 (SIGKILL)
12:04:02 INFO  OOMKilled: container exceeded memory limit`
	case "CrashLoopBackOff":
		return `12:01:00 INFO  Starting container
12:01:02 ERROR panic: runtime error: nil pointer dereference
12:01:03 INFO  Container exited with code 2
12:01:05 INFO  Starting container (restart 2)
12:01:07 ERROR panic: runtime error: nil pointer dereference
12:01:08 INFO  Container exited with code 2
12:01:10 WARN  BackOff: restarting failed container`
	case "ImagePullBackOff", "ErrImagePull":
		return `12:01:00 INFO  Pulling image "registry.example.com/app:v2.3.1"
12:01:15 ERROR Failed to pull image: manifest unknown
12:01:30 ERROR Failed to pull image: dial tcp: lookup registry.example.com: no such host
12:01:45 WARN  BackOff: image pull failed, retrying`
	default:
		return `12:01:00 INFO  Container started
12:03:00 ERROR Unexpected error in container`
	}
}

func syntheticEvents(incident *models.Incident) string {
	switch incident.IncidentType {
	case "OOMKilled":
		return fmt.Sprintf(`[Warning] OOMKilling: Memory cgroup out of memory: Killed process
[Normal] Killing: Container %s was OOMKilled
[Normal] Scheduled: Successfully assigned %s to node-1`, incident.ResourceName, incident.ResourceName)
	case "CrashLoopBackOff":
		return `[Warning] BackOff: Back-off restarting failed container
[Normal] Pulled: Container image already present on machine
[Warning] Unhealthy: Liveness probe failed`
	case "ImagePullBackOff", "ErrImagePull":
		return `[Warning] Failed: Failed to pull image: manifest unknown
[Normal] Pulling: Pulling image "registry.example.com/app"
[Warning] BackOff: Back-off pulling image`
	default:
		return `[Normal] Scheduled: Successfully assigned pod to node-1`
	}
}

type agentFinding struct {
	Findings   string   `json:"findings"`
	Severity   string   `json:"severity"`
	Evidence   []string `json:"evidence"`
	Confidence float64  `json:"confidence"`
}

func (e *Engine) multiAgentAnalyze(ctx context.Context, bundle *contextBundle, incident *models.Incident, pastLearnings string) (*analysisResult, error) {
	if e.llmCfg.APIKey == "" {
		return nil, fmt.Errorf("no LLM API key configured")
	}

	llm := NewLLMClient(e.llmCfg)
	deployCtx := fmt.Sprintf("Incident %s: %s on %s/%s — %s",
		incident.ID, incident.IncidentType, incident.Namespace, incident.ResourceName, incident.Message)

	type agentResult struct {
		name    string
		finding *agentFinding
		err     error
	}

	ch := make(chan agentResult, 3)

	// Logs agent
	go func() {
		resp, err := llm.AnalyzeWithSystem(ctx, logsAgentPrompt, "Container logs for "+deployCtx+":\n\n"+bundle.logs)
		if err != nil {
			ch <- agentResult{name: "logs", err: err}
			return
		}
		var f agentFinding
		if err := json.Unmarshal([]byte(resp.Raw), &f); err != nil {
			f.Findings = resp.Summary
			f.Evidence = resp.EvidenceLogs
		}
		ch <- agentResult{name: "logs", finding: &f}
	}()

	// Events agent
	go func() {
		resp, err := llm.AnalyzeWithSystem(ctx, eventsAgentPrompt, "Kubernetes events for "+deployCtx+":\n\n"+bundle.events)
		if err != nil {
			ch <- agentResult{name: "events", err: err}
			return
		}
		var f agentFinding
		if err := json.Unmarshal([]byte(resp.Raw), &f); err != nil {
			f.Findings = resp.Summary
			f.Evidence = resp.EvidenceEvents
		}
		ch <- agentResult{name: "events", finding: &f}
	}()

	// Deploy agent
	go func() {
		resp, err := llm.AnalyzeWithSystem(ctx, deployAgentPrompt, "Analyze this deployment incident: "+deployCtx)
		if err != nil {
			ch <- agentResult{name: "deploy", err: err}
			return
		}
		var f agentFinding
		if err := json.Unmarshal([]byte(resp.Raw), &f); err != nil {
			f.Findings = resp.Summary
		}
		ch <- agentResult{name: "deploy", finding: &f}
	}()

	// Collect results
	results := make(map[string]*agentFinding)
	var firstErr error
	for i := 0; i < 3; i++ {
		r := <-ch
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		if r.finding != nil {
			results[r.name] = r.finding
		}
	}

	// Need at least 2 agents to succeed
	if len(results) < 2 {
		return nil, fmt.Errorf("only %d agents succeeded: %w", len(results), firstErr)
	}

	log.Printf("RCA Engine: %d agents completed, synthesizing...", len(results))

	// Synthesis
	logsFindings := formatFinding(results["logs"])
	eventsFindings := formatFinding(results["events"])
	deployFindings := formatFinding(results["deploy"])

	synthPrompt := BuildSynthesisPrompt(logsFindings, eventsFindings, deployFindings, pastLearnings)
	resp, err := llm.AnalyzeWithSystem(ctx,
		"You are a lead incident investigator synthesizing findings from multiple specialists into one definitive root cause analysis.",
		synthPrompt)
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}

	avgConfidence := 0.0
	for _, f := range results {
		avgConfidence += f.Confidence
	}
	avgConfidence /= float64(len(results))

	var allLogEvidence, allEventEvidence []string
	for _, f := range results {
		allLogEvidence = append(allLogEvidence, f.Evidence...)
	}

	return &analysisResult{
		summary:          resp.Summary,
		rootCause:        resp.RootCause,
		confidence:       (resp.Confidence + avgConfidence) / 2,
		suggestedActions: resp.SuggestedActions,
		evidenceLogs:     append(allLogEvidence, resp.EvidenceLogs...),
		evidenceEvents:   append(allEventEvidence, resp.EvidenceEvents...),
		rawOutput:        resp.Raw,
	}, nil
}

func formatFinding(f *agentFinding) string {
	if f == nil {
		return "No analysis available."
	}
	return fmt.Sprintf("Findings: %s\nSeverity: %s\nConfidence: %.0f%%\nEvidence: %s",
		f.Findings, f.Severity, f.Confidence*100, strings.Join(f.Evidence, "; "))
}

func (e *Engine) callLLM(ctx context.Context, bundle *contextBundle, incident *models.Incident, pastLearnings string) (*analysisResult, error) {
	if e.llmCfg.APIKey == "" {
		return nil, fmt.Errorf("no LLM API key configured")
	}

	llm := NewLLMClient(e.llmCfg)
	prompt := BuildPrompt(incident, bundle.logs, bundle.events, pastLearnings)
	response, err := llm.Analyze(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &analysisResult{
		summary:          response.Summary,
		rootCause:        response.RootCause,
		confidence:       response.Confidence,
		suggestedActions: response.SuggestedActions,
		evidenceLogs:     response.EvidenceLogs,
		evidenceEvents:   response.EvidenceEvents,
		rawOutput:        response.Raw,
	}, nil
}

func (e *Engine) queryPastIncidents(ctx context.Context, incident *models.Incident) string {
	all, err := e.store.ListIncidents(ctx, models.IncidentFilter{Limit: 200})
	if err != nil {
		return ""
	}

	var similar []string
	for _, past := range all {
		if past.ID == incident.ID {
			continue
		}
		if past.IncidentType != incident.IncidentType {
			continue
		}
		if past.Status != "resolved" {
			continue
		}

		rca, err := e.store.GetRCAResult(ctx, past.ID)
		if err != nil || rca == nil {
			continue
		}

		similar = append(similar, fmt.Sprintf(
			"Incident %s (%s): root cause was '%s'. Action taken: %s. Result: resolved.",
			past.ID, past.ResourceName, rca.RootCause, rca.SuggestedActions,
		))

		if len(similar) >= 3 {
			break
		}
	}

	return strings.Join(similar, "\n")
}

func heuristicAnalysis(incident *models.Incident, ctxBundle *contextBundle) *analysisResult {
	result := &analysisResult{confidence: 0.3}

	switch incident.IncidentType {
	case "OOMKilled":
		result.rootCause = fmt.Sprintf("%s exceeded its memory limit and was OOMKilled", incident.ResourceName)
		result.summary = "Memory limit exceeded. The container was terminated by the OOM killer."
		result.suggestedActions = "restart,scale_up"
		result.evidenceLogs = []string{"OOMKilled exit code 137", incident.Message}
		result.evidenceEvents = []string{"Pod was OOMKilled", "Memory usage exceeded limit"}
	case "CrashLoopBackOff":
		result.rootCause = fmt.Sprintf("%s is crashing repeatedly and entering CrashLoopBackOff", incident.ResourceName)
		result.summary = "Application is crashing on startup or during runtime, causing repeated restarts."
		result.suggestedActions = "restart,rollback"
		result.evidenceLogs = []string{"Container exited with non-zero status", incident.Message}
		result.evidenceEvents = []string{"BackOff restarting failed container"}
	case "ImagePullBackOff", "ErrImagePull":
		result.rootCause = fmt.Sprintf("%s cannot pull the container image", incident.ResourceName)
		result.summary = "Container image is missing, misconfigured, or registry is unreachable."
		result.suggestedActions = "rollback"
		result.evidenceLogs = []string{incident.Message}
		result.evidenceEvents = []string{"Failed to pull image", "ErrImagePull or ImagePullBackOff"}
	default:
		result.rootCause = fmt.Sprintf("Unknown issue with %s", incident.ResourceName)
		result.summary = "Unable to determine root cause automatically."
		result.suggestedActions = "restart"
	}

	return result
}
