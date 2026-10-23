package remediation

import (
	"context"
	"testing"

	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

func newTestGate(t *testing.T, maxPerHour int) (*SafetyGate, models.Store) {
	t.Helper()
	store, err := models.NewSQLiteStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	return NewSafetyGate(kube.NewFakeClient(), store, []string{"restart", "scale_up", "rollback"}, maxPerHour), store
}

func TestSafetyGateWhitelist(t *testing.T) {
	g, _ := newTestGate(t, 20)

	rem := &models.Remediation{Type: "drain_node", Namespace: "default", TargetName: "web"}
	if err := g.Check(context.Background(), rem); err == nil {
		t.Fatal("expected block for non-whitelisted action")
	}
}

func TestSafetyGateScalePreCheck(t *testing.T) {
	g, _ := newTestGate(t, 20)

	// scale_up against a deployment that does not exist must be blocked.
	rem := &models.Remediation{Type: iforge.RemediationTypeScaleUp, Namespace: "default", TargetName: "web"}
	if err := g.Check(context.Background(), rem); err == nil {
		t.Fatal("expected scale_up pre-check to block on missing deployment")
	}

	// restart is advisory: missing pod must NOT block.
	rem = &models.Remediation{Type: iforge.RemediationTypeRestart, Namespace: "default", TargetName: "p1"}
	if err := g.Check(context.Background(), rem); err != nil {
		t.Fatalf("expected restart to pass with missing pod: %v", err)
	}
}

func TestSafetyGateCircuitBreaker(t *testing.T) {
	g, store := newTestGate(t, 3)

	for i := 0; i < 3; i++ {
		rem := &models.Remediation{
			IncidentID: "INC-1",
			Type:       "restart",
			Status:     "success",
			Namespace:  "default",
		}
		if err := store.CreateRemediation(context.Background(), rem); err != nil {
			t.Fatal(err)
		}
	}

	next := &models.Remediation{Type: "restart", Namespace: "default", TargetName: "p1"}
	if err := g.Check(context.Background(), next); err == nil {
		t.Fatal("expected circuit breaker to open at limit")
	}
}
