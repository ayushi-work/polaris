package rca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ayushi/polaris/internal/config"
)

type LLMClient struct {
	cfg        config.LLMConfig
	httpClient *http.Client
}

type LLMResponse struct {
	Summary          string   `json:"summary"`
	RootCause        string   `json:"root_cause"`
	Confidence       float64  `json:"confidence"`
	SuggestedActions string   `json:"suggested_actions"`
	EvidenceLogs     []string `json:"evidence_logs"`
	EvidenceEvents   []string `json:"evidence_events"`
	Raw              string   `json:"-"`
}

func NewLLMClient(cfg config.LLMConfig) *LLMClient {
	return &LLMClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// AnalyzeWithSystem sends a prompt with a custom system message.
func (c *LLMClient) AnalyzeWithSystem(ctx context.Context, system, prompt string) (*LLMResponse, error) {
	baseURL := c.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	body := map[string]interface{}{
		"model": c.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": prompt},
		},
		"temperature": c.cfg.Temperature,
		"max_tokens":  c.cfg.MaxTokens,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("llm returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode llm response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in llm response")
	}

	content := result.Choices[0].Message.Content

	var analysis LLMResponse
	analysis.Raw = content
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		analysis.Summary = content
		analysis.RootCause = "See summary"
	}

	return &analysis, nil
}

func (c *LLMClient) Analyze(ctx context.Context, prompt string) (*LLMResponse, error) {
	return c.AnalyzeWithSystem(ctx, systemPrompt, prompt)
}

// --- Tool-calling support (OpenAI-compatible chat completions) ---

type llmMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type toolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function funcCall `json:"function"`
}

type funcCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type toolSpec struct {
	Type     string   `json:"type"`
	Function toolFunc `json:"function"`
}

type toolFunc struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// agentTools is the toolset given to the RCA investigation agent.
var agentTools = []toolSpec{
	{
		Type: "function",
		Function: toolFunc{
			Name:        "get_pod_logs",
			Description: "Tail container logs for a pod in the incident namespace.",
			Parameters: obj(map[string]interface{}{
				"pod":   strParam("Pod name; defaults to the incident's pod"),
				"lines": intParam("Tail lines (default 80)"),
			}),
		},
	},
	{
		Type: "function",
		Function: toolFunc{
			Name:        "get_k8s_events",
			Description: "Fetch recent Kubernetes events in the incident namespace.",
			Parameters:  obj(map[string]interface{}{}),
		},
	},
	{
		Type: "function",
		Function: toolFunc{
			Name:        "get_deployment_status",
			Description: "Describe deployment replicas, rollout conditions and status.",
			Parameters: obj(map[string]interface{}{
				"pod": strParam("Workload/pod name to resolve the owning deployment"),
			}),
		},
	},
	{
		Type: "function",
		Function: toolFunc{
			Name:        "get_pod_metrics",
			Description: "Fetch 30m Prometheus memory/CPU history for the pod (empty if prometheus not configured).",
			Parameters: obj(map[string]interface{}{
				"pod": strParam("Pod name; defaults to the incident's pod"),
			}),
		},
	},
	{
		Type: "function",
		Function: toolFunc{
			Name:        "search_similar_incidents",
			Description: "Search resolved past incidents of a type with their RCA conclusions, for precedent.",
			Parameters: obj(map[string]interface{}{
				"type": strParam("Incident type filter, defaults to current incident type"),
			}),
		},
	},
}

func obj(fields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": fields}
}

func strParam(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": desc}
}

func intParam(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": desc}
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content   string     `json:"content"`
			ToolCalls []toolCall `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
}

// ChatWithMessages sends a full conversation (messages accumulate across
// agent turns) with a toolset, and returns the assistant message.
func (c *LLMClient) ChatWithMessages(ctx context.Context, messages []llmMessage, tools []toolSpec) (*chatChoiceResult, error) {
	baseURL := c.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	body := map[string]interface{}{
		"model":       c.cfg.Model,
		"messages":    messages,
		"temperature": c.cfg.Temperature,
		"max_tokens":  c.cfg.MaxTokens,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("llm returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode llm response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in llm response")
	}

	msg := result.Choices[0].Message
	return &chatChoiceResult{
		Content:   msg.Content,
		ToolCalls: msg.ToolCalls,
		Raw:       msg.Content,
	}, nil
}

type chatChoiceResult struct {
	Content   string
	ToolCalls []toolCall
	Raw       string
}

const systemPrompt = `You are a Kubernetes incident analyst. For every diagnosis, you MUST cite specific evidence from the provided logs and events. Copy exact log lines and event messages that support your conclusion. Do not guess — if the evidence isn't there, say so.`
