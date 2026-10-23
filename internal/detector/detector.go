package detector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

type Rule interface {
	Name() string
	Evaluate(ctx context.Context, client kube.Client, namespace string) ([]*models.Incident, error)
}

type Detector struct {
	client    kube.Client
	bus       eventbus.EventBus
	rules     []Rule
	interval  time.Duration
	namespace string

	lastFired map[string]time.Time
	mu        sync.Mutex
	cooldown  time.Duration
}

func New(client kube.Client, bus eventbus.EventBus, interval time.Duration, namespace string) *Detector {
	return &Detector{
		client:    client,
		bus:       bus,
		interval:  interval,
		namespace: namespace,
		lastFired: make(map[string]time.Time),
		cooldown:  5 * time.Minute,
	}
}

// key identifies an incident for dedup purposes.
func (d *Detector) key(inc *models.Incident) string {
	return inc.Namespace + "/" + inc.ResourceName + "/" + inc.IncidentType
}

func (d *Detector) shouldSuppress(inc *models.Incident) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.lastFired[d.key(inc)]; ok && time.Since(t) < d.cooldown {
		return true
	}
	return false
}

func (d *Detector) RegisterRule(rule Rule) {
	d.rules = append(d.rules, rule)
}

func (d *Detector) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	log.Printf("Detector: starting with %d rules, interval=%s", len(d.rules), d.interval)

	d.runOnce(ctx)

	for {
		select {
		case <-ticker.C:
			d.runOnce(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Detector) runOnce(ctx context.Context) {
	for _, rule := range d.rules {
		incidents, err := rule.Evaluate(ctx, d.client, d.namespace)
		if err != nil {
			log.Printf("Detector: rule %s error: %v", rule.Name(), err)
			continue
		}

		for _, inc := range incidents {
			// Dedup on the detector side: same pod + type inside cooldown
			// must not re-fire every scan.
			if d.shouldSuppress(inc) {
				continue
			}
			log.Printf("Detector: %s fired: %s in %s/%s", rule.Name(), inc.IncidentType, inc.Namespace, inc.ResourceName)
			d.lastFired[d.key(inc)] = time.Now()
			d.bus.Publish(eventbus.Event{
				// incident.detected: not yet persisted; the orchestrator
				// owns persistence for detector-sourced incidents.
				Type:    eventbus.EventIncidentDetected,
				Payload: inc,
			})
		}
	}
}

func PodHasStatus(podStatus, reason string) bool {
	return podStatus == reason
}

func NewIncident(namespace, kind, resourceName, incidentType, severity, message string) *models.Incident {
	return &models.Incident{
		Cluster:      "local",
		Namespace:    namespace,
		Kind:         kind,
		ResourceName: resourceName,
		IncidentType: incidentType,
		Severity:     severity,
		Status:       iforge.IncidentStatusDetected,
		Message:      message,
		DetectedAt:   time.Now().UTC(),
	}
}
