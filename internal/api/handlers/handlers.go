package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"

	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/pkg/iforge"
)

type Handlers struct {
	store models.Store
	bus   eventbus.EventBus
}

func New(store models.Store, bus eventbus.EventBus) *Handlers {
	return &Handlers{store: store, bus: bus}
}

// --- Health ---

func (h *Handlers) Healthz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handlers) Readyz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ready"})
}

// --- Incidents ---

func (h *Handlers) ListIncidents(c *fiber.Ctx) error {
	filter := models.IncidentFilter{
		Status:   c.Query("status"),
		Severity: c.Query("severity"),
		Service:  c.Query("service"),
		Limit:    c.QueryInt("limit", 50),
		Offset:   c.QueryInt("offset", 0),
	}
	incidents, err := h.store.ListIncidents(c.Context(), filter)
	if err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.JSON(incidents)
}

func (h *Handlers) CreateIncident(c *fiber.Ctx) error {
	var inc models.Incident
	if err := c.BodyParser(&inc); err != nil {
		return c.Status(400).JSON(iforge.ErrInvalidInput)
	}

	if err := h.store.CreateIncident(c.Context(), &inc); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	h.bus.Publish(eventbus.Event{
		Type:    eventbus.EventIncidentCreated,
		Payload: &inc,
	})

	return c.Status(201).JSON(inc)
}

func (h *Handlers) GetIncident(c *fiber.Ctx) error {
	id := c.Params("id")
	inc, err := h.store.GetIncident(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrIncidentNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.JSON(inc)
}

func (h *Handlers) AcknowledgeIncident(c *fiber.Ctx) error {
	id := c.Params("id")
	inc, err := h.store.GetIncident(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrIncidentNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	inc.Status = iforge.IncidentStatusInvestigating
	if err := h.store.UpdateIncident(c.Context(), inc); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	h.bus.Publish(eventbus.Event{
		Type:    eventbus.EventIncidentUpdated,
		Payload: inc,
	})

	return c.JSON(inc)
}

func (h *Handlers) ResolveIncident(c *fiber.Ctx) error {
	id := c.Params("id")
	inc, err := h.store.GetIncident(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrIncidentNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	now := time.Now().UTC()
	inc.Status = iforge.IncidentStatusResolved
	inc.ResolvedAt = &now
	if err := h.store.UpdateIncident(c.Context(), inc); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	h.bus.Publish(eventbus.Event{
		Type:    eventbus.EventIncidentUpdated,
		Payload: inc,
	})

	return c.JSON(inc)
}

func (h *Handlers) DeleteIncident(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.store.DeleteIncident(c.Context(), id); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.SendStatus(204)
}

func (h *Handlers) GetIncidentTimeline(c *fiber.Ctx) error {
	id := c.Params("id")
	inc, err := h.store.GetIncident(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrIncidentNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	type timelineEntry struct {
		Timestamp time.Time `json:"timestamp"`
		Event     string    `json:"event"`
		Details   string    `json:"details"`
	}

	var timeline []timelineEntry

	timeline = append(timeline, timelineEntry{
		Timestamp: inc.DetectedAt,
		Event:     "Incident Detected",
		Details:   inc.Message,
	})

	for _, rem := range inc.Remediations {
		entry := timelineEntry{
			Timestamp: rem.CreatedAt,
			Event:     "Remediation " + rem.Status,
			Details:   string(rem.Type),
		}
		if rem.ExecutedAt != nil {
			entry.Timestamp = *rem.ExecutedAt
		}
		timeline = append(timeline, entry)
	}

	if inc.RCAResult != nil {
		timeline = append(timeline, timelineEntry{
			Timestamp: inc.RCAResult.CreatedAt,
			Event:     "RCA Completed",
			Details:   inc.RCAResult.Summary,
		})
	}

	if inc.ResolvedAt != nil {
		timeline = append(timeline, timelineEntry{
			Timestamp: *inc.ResolvedAt,
			Event:     "Incident Resolved",
			Details:   "Service restored",
		})
	}

	return c.JSON(timeline)
}

// --- Remediations ---

func (h *Handlers) ListRemediations(c *fiber.Ctx) error {
	incidentID := c.Query("incident_id")
	rems, err := h.store.ListRemediations(c.Context(), incidentID)
	if err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.JSON(rems)
}

func (h *Handlers) GetRemediation(c *fiber.Ctx) error {
	id := c.Params("id")
	rem, err := h.store.GetRemediation(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrRemediationNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.JSON(rem)
}

func (h *Handlers) ApproveRemediation(c *fiber.Ctx) error {
	id := c.Params("id")
	rem, err := h.store.GetRemediation(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrRemediationNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	now := time.Now().UTC()
	rem.Status = iforge.RemediationRunning
	rem.ExecutedAt = &now

	if err := h.store.UpdateRemediation(c.Context(), rem); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	h.bus.Publish(eventbus.Event{
		Type:    eventbus.EventRemediationStarted,
		Payload: rem,
	})

	return c.JSON(rem)
}

func (h *Handlers) ExecuteRemediation(c *fiber.Ctx) error {
	return h.ApproveRemediation(c)
}

func (h *Handlers) CancelRemediation(c *fiber.Ctx) error {
	id := c.Params("id")
	rem, err := h.store.GetRemediation(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrRemediationNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	rem.Status = iforge.RemediationSkipped
	now := time.Now().UTC()
	rem.CompletedAt = &now

	if err := h.store.UpdateRemediation(c.Context(), rem); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	return c.JSON(rem)
}

// --- RCA ---

func (h *Handlers) GetRCAResult(c *fiber.Ctx) error {
	incidentID := c.Params("incident_id")
	rca, err := h.store.GetRCAResult(c.Context(), incidentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrIncidentNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.JSON(rca)
}

func (h *Handlers) TriggerRCA(c *fiber.Ctx) error {
	incidentID := c.Params("incident_id")

	existing, _ := h.store.GetRCAResult(c.Context(), incidentID)
	if existing != nil {
		return c.Status(409).JSON(iforge.ErrRCAAlreadyExists)
	}

	h.bus.Publish(eventbus.Event{
		Type: eventbus.EventRCATriggered,
		Payload: map[string]string{
			"incident_id": incidentID,
		},
	})

	return c.Status(202).JSON(fiber.Map{"status": "analysis_triggered", "incident_id": incidentID})
}

// --- Chaos ---

func (h *Handlers) ChaosStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"active_scenarios": 0, "cluster_health": "healthy"})
}

func (h *Handlers) ListChaosScenarios(c *fiber.Ctx) error {
	scenarios, err := h.store.ListChaosScenarios(c.Context())
	if err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.JSON(scenarios)
}

func (h *Handlers) CreateChaosScenario(c *fiber.Ctx) error {
	var sc models.ChaosScenario
	if err := c.BodyParser(&sc); err != nil {
		return c.Status(400).JSON(iforge.ErrInvalidInput)
	}

	if err := h.store.CreateChaosScenario(c.Context(), &sc); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	return c.Status(201).JSON(sc)
}

func (h *Handlers) GetChaosScenario(c *fiber.Ctx) error {
	id := c.Params("id")
	sc, err := h.store.GetChaosScenario(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrScenarioNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.JSON(sc)
}

func (h *Handlers) DeleteChaosScenario(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.store.DeleteChaosScenario(c.Context(), id); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}
	return c.SendStatus(204)
}

func (h *Handlers) ExecuteChaosScenario(c *fiber.Ctx) error {
	id := c.Params("id")
	sc, err := h.store.GetChaosScenario(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrScenarioNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	sc.RunCount++
	now := time.Now().UTC()
	sc.LastRunAt = &now
	_ = h.store.UpdateChaosScenario(c.Context(), sc)

	h.bus.Publish(eventbus.Event{
		Type:    eventbus.EventChaosStarted,
		Payload: sc,
	})

	return c.JSON(fiber.Map{"status": "executed", "scenario_id": id})
}

func (h *Handlers) ScheduleChaosScenario(c *fiber.Ctx) error {
	id := c.Params("id")
	sc, err := h.store.GetChaosScenario(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(iforge.ErrScenarioNotFound)
		}
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	var body struct {
		Schedule string `json:"schedule"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(iforge.ErrInvalidInput)
	}

	sc.Schedule = body.Schedule
	if err := h.store.UpdateChaosScenario(c.Context(), sc); err != nil {
		return c.Status(500).JSON(iforge.ErrInternal)
	}

	return c.JSON(sc)
}

// --- WebSocket Hub ---

type WebSocketHub struct {
	bus       eventbus.EventBus
	clients   map[*websocket.Conn]bool
	mu        sync.RWMutex
	register  chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewWebSocketHub(bus eventbus.EventBus) *WebSocketHub {
	return &WebSocketHub{
		bus:        bus,
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *WebSocketHub) Run() {
	events, unsub := h.bus.Subscribe()
	defer unsub()

	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()

		case event := <-events:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			h.mu.RLock()
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("ws write error: %v", err)
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *WebSocketHub) HandleUpgrade(c *websocket.Conn) {
	h.register <- c

	defer func() {
		h.unregister <- c
	}()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}
