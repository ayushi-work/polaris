package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"

	"github.com/ayushi/polaris/internal/api/handlers"
	"github.com/ayushi/polaris/internal/api/middleware"
	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/models"
)

func NewRouter(store models.Store, bus eventbus.EventBus) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:   "Polaris",
		BodyLimit: 4 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(middleware.Logger())

	h := handlers.New(store, bus)

	wsHub := handlers.NewWebSocketHub(bus)
	go wsHub.Run()

	api := app.Group("/api/v1")

	api.Get("/healthz", h.Healthz)
	api.Get("/readyz", h.Readyz)
	api.Get("/ws", websocket.New(wsHub.HandleUpgrade))

	incidents := api.Group("/incidents")
	incidents.Get("/", h.ListIncidents)
	incidents.Post("/", h.CreateIncident)
	incidents.Get("/:id", h.GetIncident)
	incidents.Put("/:id/acknowledge", h.AcknowledgeIncident)
	incidents.Put("/:id/resolve", h.ResolveIncident)
	incidents.Delete("/:id", h.DeleteIncident)
	incidents.Get("/:id/timeline", h.GetIncidentTimeline)

	remediations := api.Group("/remediations")
	remediations.Get("/", h.ListRemediations)
	remediations.Get("/:id", h.GetRemediation)
	remediations.Post("/:id/approve", h.ApproveRemediation)
	remediations.Post("/:id/execute", h.ExecuteRemediation)
	remediations.Post("/:id/cancel", h.CancelRemediation)

	analysis := api.Group("/analysis")
	analysis.Get("/:incident_id", h.GetRCAResult)
	analysis.Post("/:incident_id", h.TriggerRCA)

	chaos := api.Group("/chaos")
	chaos.Get("/status", h.ChaosStatus)
	scenarios := chaos.Group("/scenarios")
	scenarios.Get("/", h.ListChaosScenarios)
	scenarios.Post("/", h.CreateChaosScenario)
	scenarios.Get("/:id", h.GetChaosScenario)
	scenarios.Delete("/:id", h.DeleteChaosScenario)
	scenarios.Post("/:id/execute", h.ExecuteChaosScenario)
	scenarios.Put("/:id/schedule", h.ScheduleChaosScenario)

	return app
}
