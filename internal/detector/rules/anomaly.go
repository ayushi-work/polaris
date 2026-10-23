package rules

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayushi/polaris/internal/detector"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

// AnomalyRule is a stateful statistical rule: it tracks each container's
// restart count across detector scans and fires when the restart rate is a
// significant outlier against the pod's own baseline (z-score >= 2.5) or growth
// exceeds an absolute threshold. This catches degenerating pods even when no
// single scan shows a terminal condition like CrashLoopBackOff.
type AnomalyRule struct {
	mu       sync.Mutex
	stats    map[string]*podRestartStats
	zScore   float64
	absDelta int32
	cooldown time.Duration
}

type podRestartStats struct {
	seen           bool
	last           int32
	deltas         []float64
	sum            float64
	sumSq          float64
	n              float64
	lastFired      time.Time
	lastIncidentID string
}

func NewAnomalyRule() *AnomalyRule {
	return &AnomalyRule{
		stats:    make(map[string]*podRestartStats),
		zScore:   2.5,
		absDelta: 3,
		cooldown: 10 * time.Minute,
	}
}

func (r *AnomalyRule) Name() string { return "anomaly-restarts" }

func (r *AnomalyRule) Evaluate(ctx context.Context, client kube.Client, namespace string) ([]*models.Incident, error) {
	pods, err := client.ListPods(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var incidents []*models.Incident
	for _, pod := range pods.Items {
		var totalRestarts int32
		for _, cs := range pod.Status.ContainerStatuses {
			totalRestarts += cs.RestartCount
		}

		key := pod.Namespace + "/" + pod.Name
		st, ok := r.stats[key]
		if !ok {
			st = &podRestartStats{}
			r.stats[key] = st
		}

		if !st.seen {
			st.seen = true
			st.last = totalRestarts
			continue
		}

		delta := totalRestarts - st.last
		st.last = totalRestarts
		if delta <= 0 {
			continue
		}

		d := float64(delta)
		mean, std := st.baseline()
		if st.n >= 5 && std > 0 {
			z := (d - mean) / std
			if z >= r.zScore && time.Since(st.lastFired) >= r.cooldown {
				incidents = append(incidents, detector.NewIncident(
					pod.Namespace, "Pod", pod.Name,
					iforge.IncidentTypeHighRestartCount, iforge.SeverityWarning,
					fmt.Sprintf("Anomalous restart spike: +%d restarts in one scan (z=%.1f vs own baseline)", delta, z),
				))
				st.lastFired = time.Now()
			}
		}
		if delta >= r.absDelta && time.Since(st.lastFired) >= r.cooldown {
			incidents = append(incidents, detector.NewIncident(
				pod.Namespace, "Pod", pod.Name,
				iforge.IncidentTypeHighRestartCount, iforge.SeverityCritical,
				fmt.Sprintf("Restart storm: +%d restarts in one scan (threshold %d)", delta, r.absDelta),
			))
			st.lastFired = time.Now()
		}

		// Update baseline with the newest observation.
		st.deltas = append(st.deltas, d)
		if len(st.deltas) > 12 {
			st.deltas = st.deltas[len(st.deltas)-12:]
		}
		st.sum = 0
		st.sumSq = 0
		for _, v := range st.deltas {
			st.sum += v
			st.sumSq += v * v
		}
		st.n = float64(len(st.deltas))
	}

	return incidents, nil
}

// baseline returns mean and population std-dev of observed restart deltas.
func (s *podRestartStats) baseline() (mean, std float64) {
	if s.n == 0 {
		return 0, 0
	}
	mean = s.sum / s.n
	variance := s.sumSq/s.n - mean*mean
	if variance < 0 {
		variance = 0
	}
	return mean, math.Sqrt(variance)
}
