package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/models"
)

// Notifier posts incident-lifecycle alerts to a Slack-compatible webhook.
// Works with Slack incoming webhooks, Discord (with /slack suffix) and
// any JSON webhook that accepts {"text": "..."} — including generic webhooks
// which receive the full payload instead.
type Notifier struct {
	webhookURL string
	channel    string
	httpClient *http.Client
}

func New(webhookURL, channel string, timeout time.Duration) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		channel:    channel,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Run subscribes to the bus and blocks until ctx is cancelled.
func (n *Notifier) Run(ctx context.Context, bus eventbus.EventBus) error {
	events, unsub := bus.Subscribe(
		eventbus.EventIncidentCreated,
		eventbus.EventRCACompleted,
		eventbus.EventRemediationCompleted,
		eventbus.EventChaosStarted,
	)
	defer unsub()

	log.Printf("Alerting: notifier started (webhook configured, channel=%s)", n.channel)

	for {
		select {
		case e := <-events:
			n.handle(ctx, e)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (n *Notifier) handle(ctx context.Context, e eventbus.Event) {
	var (
		text string
	)
	switch e.Type {
	case eventbus.EventIncidentCreated:
		inc, ok := e.Payload.(*models.Incident)
		if !ok {
			return
		}
		text = fmt.Sprintf(":rotating_light: *[%s] %s* — %s/%s\n%s",
			inc.Severity, inc.IncidentType, inc.Namespace, inc.ResourceName, inc.Message)
	case eventbus.EventRCACompleted:
		r, ok := e.Payload.(*models.RCAResult)
		if !ok {
			return
		}
		text = fmt.Sprintf(":brain: *Root cause found* for %s (%.0f%% confidence)\n> %s",
			r.IncidentID, r.Confidence*100, r.RootCause)
	case eventbus.EventRemediationCompleted:
		rem, ok := e.Payload.(*models.Remediation)
		if !ok {
			return
		}
		status := ":white_check_mark: success"
		if rem.Status != "success" {
			status = ":x: " + rem.Status
		}
		text = fmt.Sprintf(":wrench: *Remediation %s* — %s on %s/%s\n%s",
			status, rem.Type, rem.Namespace, rem.TargetName, rem.Output)
	case eventbus.EventChaosStarted:
		sc, ok := e.Payload.(*models.ChaosScenario)
		if !ok {
			return
		}
		text = fmt.Sprintf(":bomb: *Chaos injected* — %s (%s)", sc.Name, sc.Action)
	default:
		return
	}

	if err := n.post(ctx, e.Type, text); err != nil {
		log.Printf("Alerting: post failed for %s: %v", e.Type, err)
	}
}

type slackPayload struct {
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text"`
}

func (n *Notifier) post(ctx context.Context, eventType eventbus.EventType, text string) error {
	body, err := json.Marshal(slackPayload{Channel: n.channel, Text: text})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Polaris-Event", string(eventType))

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
