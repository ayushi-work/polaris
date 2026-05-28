package rules

import (
	"context"

	v1 "k8s.io/api/core/v1"

	"github.com/ayushi/polaris/internal/detector"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

type OOMKilledRule struct{}

func (r *OOMKilledRule) Name() string { return "oomkilled" }

func (r *OOMKilledRule) Evaluate(ctx context.Context, client kube.Client, namespace string) ([]*models.Incident, error) {
	pods, err := client.ListPods(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var incidents []*models.Incident
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodFailed {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.LastTerminationState.Terminated != nil &&
					cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
					incidents = append(incidents, detector.NewIncident(
						pod.Namespace, "Pod", pod.Name,
						iforge.IncidentTypeOOMKilled, iforge.SeverityCritical,
						"Container "+cs.Name+" was OOMKilled",
					))
				}
			}
		}
	}
	return incidents, nil
}
