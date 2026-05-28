package rca

import (
	"context"
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
}

func NewEngine(kubeClient kube.Client, llmCfg config.LLMConfig) *Engine {
	return &Engine{
		kubeClient: kubeClient,
		llmCfg:     llmCfg,
	}
}

func (e *Engine) Analyze(ctx context.Context, incident *models.Incident) (*models.RCAResult, error) {
	log.Printf("RCA Engine: analyzing incident %s", incident.ID)

	ctxBundle, err := collectContext(ctx, e.kubeClient, incident)
	if err != nil {
		return nil, fmt.Errorf("collect context: %w", err)
	}

	analysis, err := e.callLLM(ctx, ctxBundle, incident)
	if err != nil {
		log.Printf("RCA Engine: LLM call failed, using heuristics: %v", err)
		analysis = heuristicAnalysis(incident, ctxBundle)
	}

	rca := &models.RCAResult{
		IncidentID:       incident.ID,
		Summary:          analysis.summary,
		RootCause:        analysis.rootCause,
		Confidence:       analysis.confidence,
		LogsSnippet:      ctxBundle.logs,
		EventsSummary:    ctxBundle.events,
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
	rawOutput        string
}

func collectContext(ctx context.Context, client kube.Client, incident *models.Incident) (*contextBundle, error) {
	bundle := &contextBundle{}

	logs, err := client.GetPodLogs(ctx, incident.Namespace, incident.ResourceName, 100)
	if err != nil {
		bundle.logs = fmt.Sprintf("(logs unavailable: %v)", err)
	} else {
		bundle.logs = logs
	}

	events, err := client.GetEvents(ctx, incident.Namespace, metav1.ListOptions{})
	if err != nil {
		bundle.events = fmt.Sprintf("(events unavailable: %v)", err)
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

func (e *Engine) callLLM(ctx context.Context, bundle *contextBundle, incident *models.Incident) (*analysisResult, error) {
	if e.llmCfg.APIKey == "" {
		return nil, fmt.Errorf("no LLM API key configured")
	}
	return nil, fmt.Errorf("LLM integration not yet implemented")
}

func heuristicAnalysis(incident *models.Incident, ctxBundle *contextBundle) *analysisResult {
	result := &analysisResult{confidence: 0.3}

	switch incident.IncidentType {
	case "OOMKilled":
		result.rootCause = fmt.Sprintf("%s exceeded its memory limit and was OOMKilled", incident.ResourceName)
		result.summary = "Memory limit exceeded. The container was terminated by the OOM killer."
		result.suggestedActions = "restart,scale_up"
	case "CrashLoopBackOff":
		result.rootCause = fmt.Sprintf("%s is crashing repeatedly and entering CrashLoopBackOff", incident.ResourceName)
		result.summary = "Application is crashing on startup or during runtime, causing repeated restarts."
		result.suggestedActions = "restart,rollback"
	case "ImagePullBackOff", "ErrImagePull":
		result.rootCause = fmt.Sprintf("%s cannot pull the container image", incident.ResourceName)
		result.summary = "Container image is missing, misconfigured, or registry is unreachable."
		result.suggestedActions = "rollback"
	default:
		result.rootCause = fmt.Sprintf("Unknown issue with %s", incident.ResourceName)
		result.summary = "Unable to determine root cause automatically."
		result.suggestedActions = "restart"
	}

	return result
}
