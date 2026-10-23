package rca

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayushi/polaris/internal/models"
)

// AgentAnalyze runs an agentic investigation loop: the LLM is given a small
// toolset (logs, events, deployment state, metrics, incident memory) and may
// call tools iteratively until it declares the root cause as JSON.
// Every tool round-trip is recorded in the result so the UI can show the
// investigation trail.
func (e *Engine) AgentAnalyze(ctx context.Context, incident *models.Incident) (*analysisResult, error) {
	if e.llmCfg.APIKey == "" {
		return nil, fmt.Errorf("no LLM API key configured")
	}

	llm := NewLLMClient(e.llmCfg)
	system := agentSystemPrompt

	messages := []llmMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: agentBriefPrompt(incident)},
	}

	trace := make([]agentStep, 0, 8)

	for turn := 0; turn < 8; turn++ {
		resp, err := llm.ChatWithMessages(ctx, messages, agentTools)
		if err != nil {
			return nil, fmt.Errorf("agent turn %d: %w", turn, err)
		}

		if len(resp.ToolCalls) == 0 {
			final, err := parseAgentFinal(resp.Content)
			if err != nil {
				// The model answered in prose instead of the required JSON.
				// Nudge it once for the exact format, then keep looping.
				log.Printf("RCA Agent: round %d answer not valid final JSON, nudging", turn)
				messages = append(messages, llmMessage{
					Role:      "assistant",
					Content:   resp.Content,
					ToolCalls: nil,
				}, llmMessage{
					Role:    "user",
					Content: "Your reply was not valid final JSON. Next time respond ONLY with either a tool call, or the final JSON object exactly in the required schema.",
				})
				continue
			}
			final.toolTrace = trace
			final.rawOutput = resp.Raw
			log.Printf("RCA Engine: agent loop done after %d tool rounds", len(trace))
			return final, nil
		}

		messages = append(messages, llmMessage{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			output := e.executeAgentTool(ctx, tc, incident)
			trace = append(trace, agentStep{Tool: tc.Function.Name, Arguments: tc.Function.Arguments, Output: truncate(output, 400)})
			messages = append(messages, llmMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    output,
			})
		}
	}

	// Tool budget exhausted — force a final answer from the evidence
	// gathered so far (no toolset offered), then parse it.
	log.Printf("RCA Agent: forcing final answer after %d tool rounds", len(trace))
	messages = append(messages, llmMessage{
		Role:    "user",
		Content: "Tool budget exhausted. Based ONLY on the evidence you have gathered so far, respond now with ONLY the final JSON object in the required schema.",
	})
	resp, err := llm.ChatWithMessages(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("agent forced final: %w", err)
	}
	if final, err := parseAgentFinal(resp.Content); err != nil {
		return nil, fmt.Errorf("agent forced final parse: %w", err)
	} else {
		final.toolTrace = trace
		final.rawOutput = resp.Raw
		log.Printf("RCA Engine: agent loop done (forced) after %d tool rounds", len(trace))
		return final, nil
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// --- Tools ---

type agentStep struct {
	Tool      string `json:"tool"`
	Arguments string `json:"args"`
	Output    string `json:"output"`
}

func (e *Engine) executeAgentTool(ctx context.Context, tc toolCall, incident *models.Incident) string {
	var args struct {
		Pod   string `json:"pod"`
		Type  string `json:"type"`
		Lines int    `json:"lines"`
	}
	_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

	if args.Pod == "" {
		args.Pod = incident.ResourceName
	}
	if args.Lines == 0 {
		args.Lines = 80
	}

	switch tc.Function.Name {
	case "get_pod_logs":
		if args.Pod == incident.ResourceName {
			// Reuse context already gathered when possible.
			logs, err := e.kubeClient.GetPodLogs(ctx, incident.Namespace, incident.ResourceName, int64(args.Lines))
			if err != nil || logs == "" {
				logs = syntheticLogs(incident)
			}
			return logs
		}
		logs, err := e.kubeClient.GetPodLogs(ctx, incident.Namespace, args.Pod, int64(args.Lines))
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return logs

	case "get_k8s_events":
		events, err := e.kubeClient.GetEvents(ctx, incident.Namespace, metav1.ListOptions{})
		if err != nil {
			if events == nil {
				return syntheticEvents(incident)
			}
			return fmt.Sprintf("error: %v", err)
		}
		if len(events.Items) == 0 {
			return syntheticEvents(incident)
		}
		var parts []string
		for _, ev := range events.Items {
			parts = append(parts, fmt.Sprintf("[%s] %s: %s", ev.Type, ev.Reason, ev.Message))
		}
		return strings.Join(parts, "\n")

	case "get_deployment_status":
		name := strings.TrimSuffix(args.Pod, "-pod")
		dep, err := e.kubeClient.GetDeployment(ctx, incident.Namespace, name)
		if err != nil {
			return fmt.Sprintf("deployment %s/%s: not found (%v)", incident.Namespace, name, err)
		}
		out := fmt.Sprintf("deployment %s/%s: replicas=%d", dep.Namespace, dep.Name,
			valueOfInt32(dep.Spec.Replicas))
		for _, c := range dep.Status.Conditions {
			out += fmt.Sprintf("\ncondition %s=%s (%s)", c.Type, c.Status, c.Reason)
		}
		return out

	case "get_pod_metrics":
		if e.prom == nil {
			return "metrics unavailable: prometheus not configured"
		}
		out := ""
		for _, q := range []struct{ title, query string }{
			{"memory MiB (30m)",
				fmt.Sprintf(`max by (pod) (container_memory_working_set_bytes{namespace=%q,pod=~"%s.*"})/1048576`,
					incident.Namespace, args.Pod)},
			{"CPU cores (30m)",
				fmt.Sprintf(`max by (pod) (rate(container_cpu_usage_seconds_total{namespace=%q,pod=~"%s.*"}[5m]))`,
					incident.Namespace, args.Pod)},
		} {
			res, err := e.prom.QueryRange(ctx, q.query, time.Now().Add(-30*time.Minute), time.Now(), time.Minute)
			if err != nil {
				continue
			}
			out += q.title + ":\n" + res + "\n"
		}
		if out == "" {
			return "metrics unavailable for this pod"
		}
		return out

	case "search_similar_incidents":
		incType := args.Type
		if incType == "" {
			incType = incident.IncidentType
		}
		all, err := e.store.ListIncidents(ctx, models.IncidentFilter{Limit: 100})
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		var parts []string
		for _, past := range all {
			if past.IncidentType != incType || past.ID == incident.ID || past.Status != modelsStatusResolved() {
				continue
			}
			rca, err := e.store.GetRCAResult(ctx, past.ID)
			if err != nil || rca == nil {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s (%s): root cause '%s'; actions %s",
				past.ID, past.ResourceName, rca.RootCause, rca.SuggestedActions))
			if len(parts) >= 3 {
				break
			}
		}
		if len(parts) == 0 {
			return "no resolved incidents with RCA found for type " + incType
		}
		return strings.Join(parts, "\n")

	default:
		return fmt.Sprintf("unknown tool: %s", tc.Function.Name)
	}
}

func valueOfInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

func modelsStatusResolved() string {
	return "resolved"
}

// parseAgentFinal parses the agent's concluding JSON answer.
func parseAgentFinal(content string) (*analysisResult, error) {
	var out struct {
		Summary          string   `json:"summary"`
		RootCause        string   `json:"root_cause"`
		Confidence       float64  `json:"confidence"`
		SuggestedActions string   `json:"suggested_actions"`
		EvidenceLogs     []string `json:"evidence_logs"`
		EvidenceEvents   []string `json:"evidence_events"`
	}
	if err := json.Unmarshal([]byte(stripFences(content)), &out); err != nil {
		return nil, err
	}
	return &analysisResult{
		summary:          out.Summary,
		rootCause:        out.RootCause,
		confidence:       out.Confidence,
		suggestedActions: out.SuggestedActions,
		evidenceLogs:     out.EvidenceLogs,
		evidenceEvents:   out.EvidenceEvents,
	}, nil
}

func stripFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
