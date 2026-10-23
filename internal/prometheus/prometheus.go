package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client queries a Prometheus-compatible HTTP API (v1/query, v1/query_range).
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

type queryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// Query runs an instant query and returns "label: value" lines.
func (c *Client) Query(ctx context.Context, query string) (string, error) {
	return c.decode(ctx, "query", url.Values{"query": []string{query}}, false)
}

// QueryRange runs a range query and returns a compact per-series history.
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (string, error) {
	v := url.Values{
		"query": []string{query},
		"start": []string{strconv.FormatInt(start.Unix(), 10)},
		"end":   []string{strconv.FormatInt(end.Unix(), 10)},
		"step":  []string{fmt.Sprintf("%ds", int(step.Seconds()))},
	}
	return c.decode(ctx, "query_range", v, true)
}

func (c *Client) decode(ctx context.Context, endpoint string, form url.Values, rangeMode bool) (string, error) {
	reqURL := fmt.Sprintf("%s/api/v1/%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("prometheus request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var qr queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return "", fmt.Errorf("decode prometheus response: %w", err)
	}
	if qr.Status != "success" {
		return "", fmt.Errorf("prometheus query failed: status=%s", qr.Status)
	}

	var sb strings.Builder
	n := 0
	for _, r := range qr.Data.Result {
		label := labelOf(r.Metric)
		if rangeMode && len(r.Values) > 0 {
			points := make([]string, 0, len(r.Values))
			for _, p := range r.Values {
				if len(p) < 2 {
					continue
				}
				points = append(points, fmt.Sprintf("%.2f", toFloat(p[1])))
			}
			sb.WriteString(fmt.Sprintf("%s: [%s]", label, strings.Join(points, ", ")))
		} else if len(r.Value) > 1 {
			sb.WriteString(fmt.Sprintf("%s: %.2f", label, toFloat(r.Value[1])))
		} else {
			continue
		}
		n++
		if n < len(qr.Data.Result) {
			sb.WriteString("\n")
		}
	}

	if n == 0 {
		return "", fmt.Errorf("no results for query")
	}
	return sb.String(), nil
}

func labelOf(m map[string]string) string {
	if pod, ok := m["pod"]; ok {
		return fmt.Sprintf("%s (container=%s)", pod, m["container"])
	}
	if job, ok := m["job"]; ok {
		return fmt.Sprintf("%s/%s", job, m["instance"])
	}
	return fmt.Sprintf("%v", m)
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	case float64:
		return t
	}
	return 0
}
