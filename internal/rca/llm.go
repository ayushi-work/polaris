package rca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
	analysis.Raw = content
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		// LLM didn't return valid JSON — use the raw text as summary
		analysis.Summary = content
		analysis.RootCause = "See summary"
	}

	return &analysis, nil
}

const systemPrompt = `You are a Kubernetes incident analyst. For every diagnosis, you MUST cite specific evidence from the provided logs and events. Copy exact log lines and event messages that support your conclusion. Do not guess — if the evidence isn't there, say so.`
