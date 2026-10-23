package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayushi/polaris/internal/config"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

type Engine struct {
	kubeClient kube.Client
	cfg        config.ChaosConfig
}

func NewEngine(kubeClient kube.Client, cfg config.ChaosConfig) *Engine {
	return &Engine{kubeClient: kubeClient, cfg: cfg}
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
	case iforge.ChaosActionStressCPU, iforge.ChaosActionStressMemory:
		return e.runStress(ctx, scenario)
	case iforge.ChaosActionNetworkDelay, iforge.ChaosActionNetworkLoss:
		return e.runNetem(ctx, scenario)
	case iforge.ChaosActionConfigFault:
		return e.runConfigFault(ctx, scenario)
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

// gateOpen decides whether dangerous (hostNetwork/privileged) chaos may run.
// It requires chaos.require_dangerous to be on AND the scenario to explicitly
// opt in via parameters.allow_dangerous.
func (e *Engine) gateOpen(sc *models.ChaosScenario) bool {
	if !e.cfg.RequireDangerous {
		return true
	}
	var p struct {
		AllowDangerous bool `json:"allow_dangerous"`
	}
	_ = json.Unmarshal([]byte(sc.Parameters), &p)
	return p.AllowDangerous
}

var (
	_ = context.Background
	_ = fmt.Sprintf
)
