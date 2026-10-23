package runbooks

import "testing"

func TestSelectBestMatch(t *testing.T) {
	reg := &Registry{books: []*Runbook{
		{Name: "generic", Match: Match{}, Steps: []Step{{Action: "restart"}}},
		{
			Name: "oom",
			Match: Match{IncidentTypes: []string{"OOMKilled"}, Severities: []string{"critical"}},
			Steps: []Step{{Action: "restart"}, {Action: "scale_up"}, {Action: "rollback"}},
		},
	}}

	got := reg.Select("OOMKilled", "critical")
	if got == nil || got.Name != "oom" {
		t.Fatalf("expected oom runbook, got %v", got)
	}

	got = reg.Select("PodPending", "critical")
	if got == nil || got.Name != "generic" {
		t.Fatalf("expected generic fallback, got %v", got)
	}
}

func TestNextStep(t *testing.T) {
	reg := &Registry{}
	rb := &Runbook{Name: "l", Steps: []Step{{Action: "restart"}, {Action: "scale_up"}}}

	step, ok := reg.NextStep(rb, 0)
	if !ok || step.Action != "scale_up" {
		t.Fatalf("expected scale_up for index 0, got %v ok=%v", step, ok)
	}
	if _, ok := reg.NextStep(rb, 1); ok {
		t.Fatal("expected exhaustion at last step")
	}
}

func TestMatchesCaseInsensitive(t *testing.T) {
	rb := &Runbook{Match: Match{IncidentTypes: []string{"CrashLoopBackOff"}}}
	if !matches(rb, "crashloopbackoff", "") {
		t.Fatal("expected case-insensitive type match")
	}
}
