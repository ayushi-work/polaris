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

type ImagePullRule struct{}

func (r *ImagePullRule) Name() string { return "imagepull" }

func (r *ImagePullRule) Evaluate(ctx context.Context, client kube.Client, namespace string) ([]*models.Incident, error) {
	pods, err := client.ListPods(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var incidents []*models.Incident
	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil &&
				(cs.State.Waiting.Reason == "ImagePullBackOff" || cs.State.Waiting.Reason == "ErrImagePull") {
				incidents = append(incidents, detector.NewIncident(
					pod.Namespace, "Pod", pod.Name,
					cs.State.Waiting.Reason, iforge.SeverityWarning,
					"Container "+cs.Name+" cannot pull image: "+cs.State.Waiting.Message,
				))
			}
		}
	}
	return incidents, nil
}

var (
	_ detector.Rule = (*ImagePullRule)(nil)
	_                = v1.Pod{}
	_                = metav1.ListOptions{}
	_                = kube.Client(nil)
	_                = iforge.IncidentTypeOOMKilled
	_                = context.Background
	_                = models.Incident{}
)
