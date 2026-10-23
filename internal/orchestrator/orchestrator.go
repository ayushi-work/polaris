package orchestrator

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ayushi/polaris/internal/chaos"
	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/internal/rca"
	"github.com/ayushi/polaris/internal/remediation"
	"github.com/ayushi/polaris/internal/runbooks"
	"github.com/ayushi/polaris/pkg/iforge"
)

type Orchestrator struct {
	store        models.Store
	bus          eventbus.EventBus
	rcaEngine    *rca.Engine
	remEngine    *remediation.Engine
	chaosEngine  *chaos.Engine
	kubeClient   kube.Client
	runbooks     *runbooks.Registry
	dedupeWindow time.Duration
}

func New(store models.Store, bus eventbus.EventBus, rcaEngine *rca.Engine, remEngine *remediation.Engine, chaosEngine *chaos.Engine, kubeClient kube.Client, runbookReg *runbooks.Registry) *Orchestrator {
	return &Orchestrator{
		store:        store,
		bus:          bus,
		rcaEngine:    rcaEngine,
		remEngine:    remEngine,
		chaosEngine:  chaosEngine,
		kubeClient:   kubeClient,
		runbooks:     runbookReg,
		dedupeWindow: 5 * time.Minute,
	}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	events, unsub := o.bus.Subscribe(
		eventbus.EventIncidentDetected,
		eventbus.EventIncidentCreated,
		eventbus.EventRCATriggered,
		eventbus.EventChaosStarted,
		eventbus.EventRemediationCompleted,
	)
	defer unsub()

	log.Println("Orchestrator: started, listening for events")

	for {
		select {
		case e := <-events:
			switch e.Type {
			case eventbus.EventIncidentDetected:
				o.handleIncidentDetected(ctx, e)

			case eventbus.EventIncidentCreated:
				o.handleIncidentCreated(ctx, e)

			case eventbus.EventRCATriggered:
				o.handleRCATriggered(ctx, e)

			case eventbus.EventChaosStarted:
				o.handleChaosStarted(ctx, e)

			case eventbus.EventRemediationCompleted:
				o.handleRemediationCompleted(ctx, e)
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleRemediationCompleted escalates runbook-driven remediations: if the
// failed step belongs to a runbook, auto-create the next step as pending.
func (o *Orchestrator) handleRemediationCompleted(ctx context.Context, e eventbus.Event) {
	rem, ok := e.Payload.(*models.Remediation)
	if !ok {
		return
	}
	if rem.Status != iforge.RemediationFailed {
		return
	}

	var meta struct {
		Runbook   string `json:"runbook"`
		StepIndex int    `json:"step_index"`
	}
	_ = json.Unmarshal([]byte(rem.Parameters), &meta)
	if meta.Runbook == "" || o.runbooks == nil {
		return
	}

	rb := o.runbooks.ByName(meta.Runbook)
	if rb == nil {
		return
	}
	next, ok := o.runbooks.NextStep(rb, meta.StepIndex)
	if !ok {
		log.Printf("Orchestrator: runbook %q exhausted for incident %s — escalation complete",
			rb.Name, rem.IncidentID)
		return
	}

	var m map[string]interface{}
	_ = json.Unmarshal([]byte(rem.Parameters), &m)
	m["step_index"] = meta.StepIndex + 1
	nextParams, _ := json.Marshal(m)

	escalation := &models.Remediation{
		IncidentID:  rem.IncidentID,
		Type:        next.Action,
		Status:      iforge.RemediationPending,
		TargetKind:  rem.TargetKind,
		TargetName:  rem.TargetName,
		Namespace:   rem.Namespace,
		Parameters:  string(nextParams),
		IsAutomated: true,
	}
	if err := o.store.CreateRemediation(ctx, escalation); err != nil {
		log.Printf("Orchestrator: failed to create escalation step: %v", err)
		return
	}
	log.Printf("Orchestrator: runbook %q escalating %s -> step %d (%s) for incident %s",
		rb.Name, rem.Type, meta.StepIndex+1, next.Action, rem.IncidentID)

	o.bus.Publish(eventbus.Event{
		Type:    eventbus.EventRemediationCreated,
		Payload: escalation,
	})
}

// handleIncidentDetected persists detector-sourced incidents (deduped) and
// then runs the same response chain as API-created incidents.
func (o *Orchestrator) handleIncidentDetected(ctx context.Context, e eventbus.Event) {
	inc, ok := e.Payload.(*models.Incident)
	if !ok {
		return
	}

	// Cross-check the database: another rule may have flagged this
	// workload moments ago (double detection within cooldown).
	recent, err := o.store.ListIncidents(ctx, models.IncidentFilter{
		Service: inc.ResourceName,
		Limit:   10,
	})
	if err == nil {
		for _, r := range recent {
			if r.IncidentType == inc.IncidentType &&
				r.Namespace == inc.Namespace &&
				r.Status != iforge.IncidentStatusResolved &&
				time.Since(r.DetectedAt) < o.dedupeWindow {
				log.Printf("Orchestrator: suppressing duplicate %s on %s (dedupe window)", inc.IncidentType, inc.ResourceName)
				return
			}
		}
	}

	if err := o.store.CreateIncident(ctx, inc); err != nil {
		log.Printf("Orchestrator: failed to persist detected incident: %v", err)
		return
	}
	log.Printf("Orchestrator: persisted detected incident %s (%s)", inc.ID, inc.IncidentType)

	o.bus.Publish(eventbus.Event{
		Type:    eventbus.EventIncidentCreated,
		Payload: inc,
	})
}

func (o *Orchestrator) handleIncidentCreated(ctx context.Context, e eventbus.Event) {
	inc, ok := e.Payload.(*models.Incident)
	if !ok {
		return
	}

	log.Printf("Orchestrator: incident created %s (%s)", inc.ID, inc.IncidentType)

	// Kick off RCA investigation for every new incident.
	o.bus.Publish(eventbus.Event{
		Type:    eventbus.EventRCATriggered,
		Payload: inc,
	})

	// Auto-create a remediation for critical incidents
	if inc.Severity == iforge.SeverityCritical {
		rem := o.initialRemediation(inc)
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

// initialRemediation picks step 0 of the best-matching runbook, falling back
// to a plain restart when no runbook matches.
func (o *Orchestrator) initialRemediation(inc *models.Incident) *models.Remediation {
	rem := &models.Remediation{
		IncidentID:  inc.ID,
		Type:        iforge.RemediationTypeRestart,
		Status:      iforge.RemediationPending,
		TargetKind:  inc.Kind,
		TargetName:  inc.ResourceName,
		Namespace:   inc.Namespace,
		IsAutomated: true,
	}

	if o.runbooks != nil {
		if rb := o.runbooks.Select(inc.IncidentType, inc.Severity); rb != nil {
			log.Printf("Orchestrator: runbook %q selected for %s (%d steps)", rb.Name, inc.ID, len(rb.Steps))
			step := rb.FirstStep()
			rem.Type = step.Action
			rem.Parameters = runbookMeta(rb.Name, step, inc.Severity)
		}
	}
	return rem
}

// runbookMeta encodes escalation state (runbook name + step index + severity)
// into the remediation parameters so failure handling knows what comes next.
func runbookMeta(name string, step runbooks.Step, severity string) string {
	meta, _ := json.Marshal(map[string]interface{}{
		"runbook":      name,
		"step_index":   0,
		"step_action":  step.Action,
		"min_replicas": step.MinReplicas,
		"severity":     severity,
	})
	return string(meta)
}

func (o *Orchestrator) handleRCATriggered(ctx context.Context, e eventbus.Event) {
	var incidentID string
	switch p := e.Payload.(type) {
	case map[string]string:
		incidentID = p["incident_id"]
	case *models.Incident:
		incidentID = p.ID
	default:
		return
	}
	log.Printf("Orchestrator: RCA triggered for %s — starting investigation...", incidentID)

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

	message := "Chaos scenario injected: " + sc.Description
	details := ""

	// Actually run the chaos action through the engine.
	if o.chaosEngine != nil {
		result := o.chaosEngine.Execute(ctx, sc)
		details = result.Output
		if !result.Success && result.Error != nil {
			details = result.Output + " (error: " + result.Error.Error() + ")"
		}
		log.Printf("Orchestrator: chaos engine result for %s: %s", sc.ID, details)
		o.bus.Publish(eventbus.Event{
			Type:    eventbus.EventChaosCompleted,
			Payload: sc,
		})
	}

	inc := &models.Incident{
		Cluster:      "local",
		Namespace:    "default",
		Kind:         "Pod",
		ResourceName: sc.Name,
		IncidentType: "ChaosInjection",
		Severity:     iforge.SeverityWarning,
		Status:       iforge.IncidentStatusDetected,
		Message:      message,
		Details:      details,
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
