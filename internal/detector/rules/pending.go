package rules

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayushi/polaris/internal/detector"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

type PodPendingRule struct{}

func (r *PodPendingRule) Name() string { return "podpending" }

func (r *PodPendingRule) Evaluate(ctx context.Context, client kube.Client, namespace string) ([]*models.Incident, error) {
	pods, err := client.ListPods(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var incidents []*models.Incident
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodPending {
			for _, cs := range pod.Status.Conditions {
				if cs.Type == v1.PodScheduled && cs.Status == v1.ConditionFalse &&
					cs.Reason == "Unschedulable" {
					incidents = append(incidents, detector.NewIncident(
						pod.Namespace, "Pod", pod.Name,
						iforge.IncidentTypePodPending, iforge.SeverityWarning,
						"Pod stuck in Pending state: "+cs.Message,
					))
				}
			}
		}
	}
	return incidents, nil
}

var _ detector.Rule = (*PodPendingRule)(nil)
