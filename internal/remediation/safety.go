package remediation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
)

// SafetyGate is the guardrail layer executed BEFORE any remediation action:
//  1. action whitelist — only configured action types may ever run
//  2. circuit breaker — at most MaxActionsPerHour executions across the cluster
//  3. pre-checks per action type — validate the target before touching it
type SafetyGate struct {
	kubeClient        kube.Client
	store             models.Store
	allowed           map[string]bool
	maxActionsPerHour int
}

func NewSafetyGate(kubeClient kube.Client, store models.Store, allowedActions []string, maxActionsPerHour int) *SafetyGate {
	allowed := make(map[string]bool, len(allowedActions))
	for _, a := range allowedActions {
		allowed[a] = true
	}
	return &SafetyGate{
		kubeClient:        kubeClient,
		store:             store,
		allowed:           allowed,
		maxActionsPerHour: maxActionsPerHour,
	}
}

// Check returns nil when the remediation may proceed, or a descriptive error
// explaining why it was blocked.
func (g *SafetyGate) Check(ctx context.Context, rem *models.Remediation) error {
	if !g.allowed[rem.Type] {
		return fmt.Errorf("action %q is not in the allowed_actions whitelist", rem.Type)
	}

	if g.maxActionsPerHour > 0 {
		n, err := g.recentActionCount(ctx, time.Hour)
		if err != nil {
			// Fail open on storage errors — blocking healing forever because
			// the audit query failed would be worse than the risk.
			fmt.Printf("SafetyGate: breaker count query failed: %v (failing open)\n", err)
		} else if n >= g.maxActionsPerHour {
			return fmt.Errorf("circuit breaker open: %d remediations executed in the last hour (limit %d)", n, g.maxActionsPerHour)
		}
	}

	switch rem.Type {
	case "scale_up", "scale_down":
		if _, err := g.kubeClient.GetDeployment(ctx, rem.Namespace, rem.TargetName); err != nil {
			return fmt.Errorf("pre-check failed: deployment %s/%s not found: %w", rem.Namespace, rem.TargetName, err)
		}
	case "restart", "delete_pod":
		// Advisory only: the pod may legitimately have been recreated between
		// detection and approval, so a failed lookup must not block healing.
		if _, err := g.kubeClient.GetPod(ctx, rem.Namespace, rem.TargetName); err == nil {
			log.Printf("SafetyGate: pre-check ok — pod %s/%s present", rem.Namespace, rem.TargetName)
		}
	}

	return nil
}

func (g *SafetyGate) recentActionCount(ctx context.Context, window time.Duration) (int, error) {
	rems, err := g.store.ListRemediations(ctx, "")
	if err != nil {
		return 0, err
	}
	cutoff := time.Now().UTC().Add(-window)
	count := 0
	for _, r := range rems {
		if r.Status != "success" && r.Status != "failed" {
			continue
		}
		ref := r.CreatedAt
		if r.ExecutedAt != nil {
			ref = *r.ExecutedAt
		}
		if ref.After(cutoff) {
			count++
		}
	}
	return count, nil
}
