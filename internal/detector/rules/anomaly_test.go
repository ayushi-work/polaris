package rules

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/pkg/iforge"
)

func TestAnomalyRuleDetectsRestartStorm(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "wobbly-pod", Namespace: "default", UID: types.UID("u1")},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	client := kube.NewFakeClient(pod)
	rule := NewAnomalyRule()
	ctx := context.Background()

	// First evaluation: baseline establishment, no incident.
	incidents, err := rule.Evaluate(ctx, client, "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(incidents) != 0 {
		t.Fatalf("expected no incidents on first scan, got %d", len(incidents))
	}

	// Simulate a restart storm: 5 restarts in the next scan.
	pod.Status.ContainerStatuses = []corev1.ContainerStatus{
		{Name: "app", RestartCount: 5},
	}
	if _, err := client.UpdatePodStatus(ctx, pod); err != nil {
		t.Fatal(err)
	}

	incidents, err = rule.Evaluate(ctx, client, "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(incidents) != 1 {
		t.Fatalf("expected 1 restart-storm incident, got %d", len(incidents))
	}
	if incidents[0].IncidentType != iforge.IncidentTypeHighRestartCount {
		t.Fatalf("expected %s, got %s", iforge.IncidentTypeHighRestartCount, incidents[0].IncidentType)
	}
}

func TestAnomalyRuleSilentWhenStable(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "calm-pod", Namespace: "default", UID: types.UID("u2")},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	client := kube.NewFakeClient(pod)
	rule := NewAnomalyRule()
	ctx := context.Background()

	for round := 0; round < 8; round++ {
		if round > 0 {
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{Name: "app", RestartCount: int32(round)}, // +1 per scan: normal churn
			}
			if _, err := client.UpdatePodStatus(ctx, pod); err != nil {
				t.Fatal(err)
			}
		}
		incidents, err := rule.Evaluate(ctx, client, "default")
		if err != nil {
			t.Fatal(err)
		}
		if len(incidents) != 0 {
			t.Fatalf("round %d: expected no incidents for steady churn, got %d", round, len(incidents))
		}
	}
}
