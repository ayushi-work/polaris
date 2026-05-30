package rca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ayushi/polaris/internal/config"
)

type LLMClient struct {
	cfg        config.LLMConfig
	httpClient *http.Client
}

type LLMResponse struct {
	Summary          string  `json:"summary"`
	RootCause        string  `json:"root_cause"`
	Confidence       float64 `json:"confidence"`
	SuggestedActions string  `json:"suggested_actions"`
	Raw              string  `json:"raw"`
}

func NewLLMClient(cfg config.LLMConfig) *LLMClient {
	return &LLMClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *LLMClient) Analyze(ctx context.Context, prompt string) (*LLMResponse, error) {
	baseURL := c.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	body := map[string]interface{}{
		"model": c.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
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
		return nil, fmt.Errorf("llm returned status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode llm response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in llm response")
	}

	content := result.Choices[0].Message.Content

	var analysis LLMResponse
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		analysis = LLMResponse{
			Summary:   content,
			RootCause: "See summary",
			Raw:       content,
		}
	}

	return &analysis, nil
}

const systemPrompt = `You are a Kubernetes incident analyst. Given pod logs, events, and metrics, determine the root cause of the incident and suggest remediation actions.

Respond ONLY in valid JSON with this structure:
{
  "summary": "One-line summary of the incident",
  "root_cause": "Detailed root cause analysis",
  "confidence": 0.0-1.0,
  "suggested_actions": "comma-separated list: restart,scale_up,scale_down,rollback,drain_node,cordon_node,delete_pod"
}`

var _ = time.Now
