package rules

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayushi/polaris/internal/detector"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

type NodePressureRule struct{}

func (r *NodePressureRule) Name() string { return "nodepressure" }

func (r *NodePressureRule) Evaluate(ctx context.Context, client kube.Client, namespace string) ([]*models.Incident, error) {
	pods, err := client.ListPods(ctx, namespace, metav1.ListOptions{
		FieldSelector: "spec.nodeName!=",
	})
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var incidents []*models.Incident

	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "NodePressure" {
				key := pod.Spec.NodeName + "-pressure"
				if !seen[key] {
					seen[key] = true
					incidents = append(incidents, detector.NewIncident(
						pod.Namespace, "Node", pod.Spec.NodeName,
						iforge.IncidentTypeNodePressure, iforge.SeverityWarning,
						"Node "+pod.Spec.NodeName+" is under resource pressure",
					))
				}
			}
		}
	}
	return incidents, nil
}

var _ detector.Rule = (*NodePressureRule)(nil)
var _ = v1.Pod{}
var _ = time.Now
