package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ayushi/polaris/internal/config"
	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/models"
)

type Server struct {
	app    *fiber.App
	cfg    config.ServerConfig
	store  models.Store
	bus    eventbus.EventBus
}

func NewServer(cfg config.ServerConfig, store models.Store, bus eventbus.EventBus) *Server {
	app := NewRouter(store, bus)
	return &Server{
		app:   app,
		cfg:   cfg,
		store: store,
		bus:   bus,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
		s.bus.Close()
	}()

	log.Printf("Polaris API server starting on %s", addr)
	return s.app.Listen(addr)
}
