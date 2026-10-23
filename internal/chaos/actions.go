package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

// runnerLabels tags everything the chaos engine creates so a cleanup job can
// find and garbage-collect runners.
func runnerLabels(scenarioID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "polaris-chaos",
		"polaris.io/chaos-runner":      "true",
		"polaris.io/chaos-scenario":    scenarioID,
	}
}

// chaosParams is the optional JSON in ChaosScenario.Parameters.
type chaosParams struct {
	Duration   string `json:"duration"`    // runner lifetime, e.g. "60s"
	MemoryMB   int    `json:"memory_mb"`   // for stress_memory
	CPUWorkers int    `json:"cpu_workers"` // for stress_cpu
	Node       string `json:"node"`        // target node for netem runner
}

func parseParams(raw string) chaosParams {
	var p chaosParams
	_ = json.Unmarshal([]byte(raw), &p)
	if p.Duration == "" {
		p.Duration = "60s"
	}
	if p.MemoryMB == 0 {
		p.MemoryMB = 64
	}
	if p.CPUWorkers == 0 {
		p.CPUWorkers = 2
	}
	return p
}

func runnerName(sc *models.ChaosScenario) string {
	return fmt.Sprintf("polaris-chaos-%s", sc.ID)
}

func (e *Engine) namespaceFor(sc *models.ChaosScenario) string {
	if len(e.cfg.AllowedNamespaces) > 0 {
		return e.cfg.AllowedNamespaces[0]
	}
	return "default"
} // stressRunnerPod creates a busybox pod that spins CPU or holds memory
// allocations for the configured duration, then exits.
func stressRunnerPod(sc *models.ChaosScenario) (*corev1.Pod, error) {
	p := parseParams(sc.Parameters)
	ttl, err := time.ParseDuration(p.Duration)
	if err != nil {
		return nil, fmt.Errorf("bad duration %q: %w", p.Duration, err)
	}
	timeoutSec := int(ttl.Seconds())

	var command string
	switch sc.Action {
	case iforge.ChaosActionStressCPU:
		command = fmt.Sprintf(
			"echo stress_cpu for %ds; n=%d; end=$(( $(date +%%s) + %d )); while [ $(date +%%s) -lt $end ]; do i=0; while [ $i -lt $n ]; do sh -c 'while :; do :; done' & i=$((i+1)); done; wait; done",
			timeoutSec, p.CPUWorkers, timeoutSec)
	case iforge.ChaosActionStressMemory:
		command = fmt.Sprintf(
			"echo stress_memory %d MiB chunks for %ds; end=$(( $(date +%%s) + %d )); while [ $(date +%%s) -lt $end ]; do A=$(dd if=/dev/zero bs=1M count=%d 2>/dev/null); sleep 1; done; echo done",
			p.MemoryMB, timeoutSec, timeoutSec, p.MemoryMB)
	default:
		return nil, fmt.Errorf("not a stress action: %s", sc.Action)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runnerName(sc),
			Namespace: "default",
			Labels:    runnerLabels(sc.ID),
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "chaos-runner",
					Image:   "busybox:1.36",
					Command: []string{"sh", "-c", command},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("256Mi"),
							corev1.ResourceCPU:    resource.MustParse("500m"),
						},
					},
				},
			},
		},
	}
	return pod, nil
}

// netemRunnerPod targets a NODE with tc netem for real network delay/loss.
// Requires hostNetwork + privileged, and is gated by chaos.require_dangerous.
func netemRunnerPod(sc *models.ChaosScenario) (*corev1.Pod, error) {
	p := parseParams(sc.Parameters)
	ttl, err := time.ParseDuration(p.Duration)
	if err != nil {
		return nil, fmt.Errorf("bad duration %q: %w", p.Duration, err)
	}

	var netemArg string
	switch sc.Action {
	case iforge.ChaosActionNetworkDelay:
		netemArg = fmt.Sprintf("tc qdisc add dev eth0 root netem delay 200ms; sleep %.0f; tc qdisc del dev eth0 root", ttl.Seconds())
	case iforge.ChaosActionNetworkLoss:
		netemArg = fmt.Sprintf("tc qdisc add dev eth0 root netem loss 10%%; sleep %.0f; tc qdisc del dev eth0 root", ttl.Seconds())
	default:
		return nil, fmt.Errorf("not a network action: %s", sc.Action)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runnerName(sc),
			Namespace: "default",
			Labels:    runnerLabels(sc.ID),
		},
		Spec: corev1.PodSpec{
			HostNetwork:                   true,
			RestartPolicy:                 corev1.RestartPolicyNever,
			TerminationGracePeriodSeconds: ptr.To(int64(1)),
			Containers: []corev1.Container{
				{
					Name:    "netem-runner",
					Image:   "alpine:3.20",
					Command: []string{"sh", "-c", fmt.Sprintf("apk add --no-cache iproute2 >/dev/null 2>&1 || true; %s", netemArg)},
					SecurityContext: &corev1.SecurityContext{
						Privileged: ptr.To(true),
					},
				},
			},
		},
	}
	if p.Node != "" {
		pod.Spec.NodeName = p.Node
	}
	return pod, nil
}

// runStress creates the runner pod and reports what it will do.
func (e *Engine) runStress(ctx context.Context, sc *models.ChaosScenario) *ExecutionResult {
	pod, err := stressRunnerPod(sc)
	if err != nil {
		return &ExecutionResult{Success: false, Output: err.Error(), Error: err}
	}
	created, err := e.kubeClient.CreatePod(ctx, pod)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to create stress runner: %v", err),
			Error:   err,
		}
	}
	return &ExecutionResult{
		Success: true,
		Output: fmt.Sprintf("Created runner pod %s/%s (%s, %s)",
			created.Namespace, created.Name, sc.Action, parseParams(sc.Parameters).Duration),
	}
}

// runNetem creates a privileged hostNetwork tc runner. Gated by require_dangerous.
func (e *Engine) runNetem(ctx context.Context, sc *models.ChaosScenario) *ExecutionResult {
	if !e.gateOpen(sc) {
		return &ExecutionResult{
			Success: false,
			Output:  "network chaos denied: requires chaos.require_dangerous=true AND parameters.allow_dangerous=true (hostNetwork tc netem)",
		}
	}
	pod, err := netemRunnerPod(sc)
	if err != nil {
		return &ExecutionResult{Success: false, Output: err.Error(), Error: err}
	}
	pod.Namespace = e.namespaceFor(sc)
	created, err := e.kubeClient.CreatePod(ctx, pod)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to create netem runner: %v", err),
			Error:   err,
		}
	}
	return &ExecutionResult{
		Success: true,
		Output: fmt.Sprintf("Created hostNetwork netem runner %s/%s (%s, %s)",
			created.Namespace, created.Name, sc.Action, parseParams(sc.Parameters).Duration),
	}
}

// runConfigFault patches the target deployment with a fault env var, forcing
// the app to fail in a controlled way. Reversible by removing the env var.
func (e *Engine) runConfigFault(ctx context.Context, sc *models.ChaosScenario) *ExecutionResult {
	if sc.Target == "" {
		return &ExecutionResult{
			Success: false,
			Output:  "config_fault requires target (deployment name) in scenario.target",
		}
	}

	dep, err := e.kubeClient.GetDeployment(ctx, e.namespaceFor(sc), sc.Target)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to get deployment %s: %v", sc.Target, err),
			Error:   err,
		}
	}

	faultVal := fmt.Sprintf("chaos:%s", sc.Action)
	for i := range dep.Spec.Template.Spec.Containers {
		c := &dep.Spec.Template.Spec.Containers[i]
		exists := false
		for j := range c.Env {
			if c.Env[j].Name == "POLARIS_CHAOS_FAULT" {
				c.Env[j].Value = faultVal
				exists = true
				break
			}
		}
		if !exists {
			c.Env = append(c.Env, corev1.EnvVar{Name: "POLARIS_CHAOS_FAULT", Value: faultVal})
		}
	}

	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations["polaris.io/chaos-fault"] = faultVal

	if _, err := e.kubeClient.UpdateDeployment(ctx, dep.Namespace, dep); err != nil {
		return &ExecutionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to patch deployment %s: %v", sc.Target, err),
			Error:   err,
		}
	}

	return &ExecutionResult{
		Success: true,
		Output: fmt.Sprintf("Injecting config fault on deployment %s: env POLARIS_CHAOS_FAULT=%q",
			sc.Target, faultVal),
	}
}
