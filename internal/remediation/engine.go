package remediation

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
	kubeClient  kube.Client
	maxRetries  int
	autoApprove bool
}

func NewEngine(kubeClient kube.Client, maxRetries int, autoApprove bool) *Engine {
	return &Engine{
		kubeClient:  kubeClient,
		maxRetries:  maxRetries,
		autoApprove: autoApprove,
	}
}

type ActionResult struct {
	Success bool
	Output  string
	Error   error
}

func (e *Engine) Execute(ctx context.Context, rem *models.Remediation) *ActionResult {
	log.Printf("Remediation Engine: executing %s (%s) on %s/%s", rem.Type, rem.ID, rem.Namespace, rem.TargetName)

	rem.Status = iforge.RemediationRunning
	now := time.Now().UTC()
	rem.ExecutedAt = &now

	switch rem.Type {
	case iforge.RemediationTypeRestart, iforge.RemediationTypeDeletePod:
		return e.executeRestart(ctx, rem)
	case iforge.RemediationTypeScaleUp, iforge.RemediationTypeScaleDown:
		return e.executeScale(ctx, rem)
	case iforge.RemediationTypeRollback:
		return e.executeRollback(ctx, rem)
	default:
		return &ActionResult{
			Success: false,
			Error:   fmt.Errorf("unknown remediation type: %s", rem.Type),
		}
	}
}

func (e *Engine) executeRestart(ctx context.Context, rem *models.Remediation) *ActionResult {
	err := e.kubeClient.DeletePod(ctx, rem.Namespace, rem.TargetName, metav1.DeleteOptions{})
	if err != nil {
		return &ActionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to delete pod: %v", err),
			Error:   err,
		}
	}
	return &ActionResult{
		Success: true,
		Output:  fmt.Sprintf("Pod %s/%s deleted successfully; will be recreated by controller", rem.Namespace, rem.TargetName),
	}
}

func (e *Engine) executeScale(ctx context.Context, rem *models.Remediation) *ActionResult {
	dep, err := e.kubeClient.GetDeployment(ctx, rem.Namespace, rem.TargetName)
	if err != nil {
		return &ActionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to get deployment: %v", err),
			Error:   err,
		}
	}

	currentReplicas := int32(1)
	if dep.Spec.Replicas != nil {
		currentReplicas = *dep.Spec.Replicas
	}

	var targetReplicas int32
	if rem.Type == iforge.RemediationTypeScaleUp {
		targetReplicas = currentReplicas + 1
	} else {
		targetReplicas = currentReplicas - 1
		if targetReplicas < 1 {
			targetReplicas = 1
		}
	}

	_, err = e.kubeClient.ScaleDeployment(ctx, rem.Namespace, rem.TargetName, targetReplicas)
	if err != nil {
		return &ActionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to scale deployment: %v", err),
			Error:   err,
		}
	}

	return &ActionResult{
		Success: true,
		Output:  fmt.Sprintf("Deployment %s/%s scaled to %d replicas", rem.Namespace, rem.TargetName, targetReplicas),
	}
}

func (e *Engine) executeRollback(ctx context.Context, rem *models.Remediation) *ActionResult {
	dep, err := e.kubeClient.RollbackDeployment(ctx, rem.Namespace, rem.TargetName)
	if err != nil {
		return &ActionResult{
			Success: false,
			Output:  fmt.Sprintf("Failed to rollback deployment: %v", err),
			Error:   err,
		}
	}

	rev := dep.Annotations["polaris/rolled-back-to"]
	from := dep.Annotations["polaris/rolled-back-from"]
	msg := fmt.Sprintf("Deployment %s/%s rolled back from revision %s to %s", rem.Namespace, rem.TargetName, from, rev)
	if rev == "" {
		msg = fmt.Sprintf("Deployment %s/%s rollback triggered", rem.Namespace, rem.TargetName)
	}

	return &ActionResult{
		Success: true,
		Output:  msg,
	}
}
