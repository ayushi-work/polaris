package orchestrator

import (
	"context"
	"log"
	"time"

	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/internal/rca"
	"github.com/ayushi/polaris/internal/remediation"
	"github.com/ayushi/polaris/pkg/iforge"
)

type Orchestrator struct {
	store      models.Store
	bus        eventbus.EventBus
	rcaEngine  *rca.Engine
	remEngine  *remediation.Engine
	kubeClient kube.Client
}

func New(store models.Store, bus eventbus.EventBus, rcaEngine *rca.Engine, remEngine *remediation.Engine, kubeClient kube.Client) *Orchestrator {
	return &Orchestrator{
		store:      store,
		bus:        bus,
		rcaEngine:  rcaEngine,
		remEngine:  remEngine,
		kubeClient: kubeClient,
	}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	events, unsub := o.bus.Subscribe(
		eventbus.EventIncidentCreated,
		eventbus.EventRCATriggered,
		eventbus.EventChaosStarted,
	)
	defer unsub()

	log.Println("Orchestrator: started, listening for events")

	for {
		select {
		case e := <-events:
			switch e.Type {
			case eventbus.EventIncidentCreated:
				o.handleIncidentCreated(ctx, e)

			case eventbus.EventRCATriggered:
				o.handleRCATriggered(ctx, e)

			case eventbus.EventChaosStarted:
				o.handleChaosStarted(ctx, e)
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *Orchestrator) handleIncidentCreated(ctx context.Context, e eventbus.Event) {
	inc, ok := e.Payload.(*models.Incident)
	if !ok {
		return
	}

	log.Printf("Orchestrator: incident created %s (%s)", inc.ID, inc.IncidentType)

	// Auto-create a remediation for critical incidents
	if inc.Severity == iforge.SeverityCritical {
		rem := &models.Remediation{
			IncidentID:  inc.ID,
			Type:        iforge.RemediationTypeRestart,
			Status:      iforge.RemediationPending,
			TargetKind:  inc.Kind,
			TargetName:  inc.ResourceName,
			Namespace:   inc.Namespace,
			IsAutomated: true,
		}
		if err := o.store.CreateRemediation(ctx, rem); err != nil {
			log.Printf("Orchestrator: failed to create remediation: %v", err)
			return
		}

		o.bus.Publish(eventbus.Event{
			Type:    eventbus.EventRemediationCreated,
			Payload: rem,
		})

		log.Printf("Orchestrator: remediation %s created for incident %s", rem.ID, inc.ID)
	}
}

func (o *Orchestrator) handleRCATriggered(ctx context.Context, e eventbus.Event) {
	payload, ok := e.Payload.(map[string]string)
	if !ok {
		return
	}

	incidentID := payload["incident_id"]
	log.Printf("Orchestrator: RCA triggered for %s — calling DeepSeek...", incidentID)

	inc, err := o.store.GetIncident(ctx, incidentID)
	if err != nil {
		log.Printf("Orchestrator: failed to get incident %s: %v", incidentID, err)
		return
	}

	inc.Status = iforge.IncidentStatusInvestigating
	_ = o.store.UpdateIncident(ctx, inc)

	o.bus.Publish(eventbus.Event{
		Type:    eventbus.EventIncidentUpdated,
		Payload: inc,
	})

	// Actually call the RCA engine (which calls the LLM)
	result, err := o.rcaEngine.Analyze(ctx, inc)
	if err != nil {
		log.Printf("Orchestrator: RCA analysis failed: %v", err)
		result = &models.RCAResult{
			IncidentID: incidentID,
			Summary:    "Analysis failed: " + err.Error(),
			RootCause:  "Unable to determine root cause",
			Confidence: 0.0,
			LLMModel:   "error",
			CreatedAt:  time.Now().UTC(),
		}
	}

	if err := o.store.CreateRCAResult(ctx, result); err != nil {
		log.Printf("Orchestrator: failed to store RCA result: %v", err)
		return
	}

	o.bus.Publish(eventbus.Event{
		Type:    eventbus.EventRCACompleted,
		Payload: result,
	})

	log.Printf("Orchestrator: RCA completed for %s (confidence: %.0f%%)", incidentID, result.Confidence*100)
}

func (o *Orchestrator) handleChaosStarted(ctx context.Context, e eventbus.Event) {
	sc, ok := e.Payload.(*models.ChaosScenario)
	if !ok {
		return
	}

	log.Printf("Orchestrator: chaos scenario started %s: %s", sc.ID, sc.Name)

	inc := &models.Incident{
		Cluster:      "local",
		Namespace:    "default",
		Kind:         "Pod",
		ResourceName: sc.Name,
		IncidentType: "ChaosInjection",
		Severity:     iforge.SeverityWarning,
		Status:       iforge.IncidentStatusDetected,
		Message:      "Chaos scenario injected: " + sc.Description,
		DetectedAt:   time.Now().UTC(),
	}

	if err := o.store.CreateIncident(ctx, inc); err != nil {
		log.Printf("Orchestrator: failed to create chaos incident: %v", err)
		return
	}

	o.bus.Publish(eventbus.Event{
		Type:    eventbus.EventIncidentCreated,
		Payload: inc,
	})
}
