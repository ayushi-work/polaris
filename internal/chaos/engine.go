package chaos

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

type Engine struct {
	kubeClient kube.Client
}

func NewEngine(kubeClient kube.Client) *Engine {
	return &Engine{kubeClient: kubeClient}
}

type ExecutionResult struct {
	Success   bool
	Output    string
	Error     error
	StartedAt time.Time
	EndedAt   time.Time
}

func (e *Engine) Execute(ctx context.Context, scenario *models.ChaosScenario) *ExecutionResult {
	log.Printf("Chaos Engine: executing scenario %s (%s)", scenario.ID, scenario.Action)

	result := &ExecutionResult{StartedAt: time.Now().UTC()}
	defer func() { result.EndedAt = time.Now().UTC() }()

	switch scenario.Action {
	case iforge.ChaosActionDeletePod:
		return e.deletePod(ctx, scenario)
	case iforge.ChaosActionStressCPU:
		return e.noopResult(scenario, "CPU stress not yet implemented (requires stress-ng DaemonSets)")
	case iforge.ChaosActionStressMemory:
		return e.noopResult(scenario, "Memory stress not yet implemented (requires stress-ng DaemonSets)")
	case iforge.ChaosActionNetworkDelay:
		return e.noopResult(scenario, "Network delay not yet implemented (requires network policies)")
	case iforge.ChaosActionNetworkLoss:
		return e.noopResult(scenario, "Network loss not yet implemented (requires network policies)")
	case iforge.ChaosActionConfigFault:
		return e.noopResult(scenario, "Config fault not yet implemented")
	default:
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("unknown chaos action: %s", scenario.Action),
			Output:  "Unknown action type",
		}
	}
}

func (e *Engine) deletePod(ctx context.Context, scenario *models.ChaosScenario) *ExecutionResult {
	pods, err := e.kubeClient.ListPods(ctx, "default", metav1.ListOptions{})
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   err,
			Output:  fmt.Sprintf("Failed to list pods: %v", err),
		}
	}

	if len(pods.Items) == 0 {
		return &ExecutionResult{
			Success: false,
			Output:  "No pods found to delete",
		}
	}

	target := pods.Items[0]
	err = e.kubeClient.DeletePod(ctx, target.Namespace, target.Name, metav1.DeleteOptions{})
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   err,
			Output:  fmt.Sprintf("Failed to delete pod %s/%s: %v", target.Namespace, target.Name, err),
		}
	}

	return &ExecutionResult{
		Success: true,
		Output:  fmt.Sprintf("Deleted pod %s/%s", target.Namespace, target.Name),
	}
}

func (e *Engine) noopResult(scenario *models.ChaosScenario, msg string) *ExecutionResult {
	return &ExecutionResult{
		Success: false,
		Output:  msg,
	}
}

var (
	_ = context.Background
	_ = fmt.Sprintf
)
