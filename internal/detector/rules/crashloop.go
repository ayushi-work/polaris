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

type CrashLoopRule struct{}

func (r *CrashLoopRule) Name() string { return "crashloop" }

func (r *CrashLoopRule) Evaluate(ctx context.Context, client kube.Client, namespace string) ([]*models.Incident, error) {
	pods, err := client.ListPods(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var incidents []*models.Incident
	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
				incidents = append(incidents, detector.NewIncident(
					pod.Namespace, "Pod", pod.Name,
					iforge.IncidentTypeCrashLoopBackOff, iforge.SeverityCritical,
					"Container "+cs.Name+" is in CrashLoopBackOff: "+cs.State.Waiting.Message,
				))
			}
		}
	}
	return incidents, nil
}

var _ detector.Rule = (*CrashLoopRule)(nil)
var _ = v1.Pod{}
