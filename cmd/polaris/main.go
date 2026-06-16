package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/ayushi/polaris/internal/api"
	"github.com/ayushi/polaris/internal/config"
	"github.com/ayushi/polaris/internal/detector"
	"github.com/ayushi/polaris/internal/detector/rules"
	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/kube"
	"github.com/ayushi/polaris/internal/models"
	"github.com/ayushi/polaris/internal/orchestrator"
	"github.com/ayushi/polaris/internal/rca"
	"github.com/ayushi/polaris/internal/remediation"
)

var (
	cfgPath string
	dryRun  bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "polaris",
		Short: "AI-powered Kubernetes incident simulation and self-healing",
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		RunE:  runServe,
	}
	serveCmd.Flags().StringVar(&cfgPath, "config", "", "Config file path")
	serveCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run in fake mode (no real cluster)")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("polaris v0.1.0")
		},
	}

	rootCmd.AddCommand(serveCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if dryRun {
		cfg.Kubernetes.Mode = "fake"
		log.Println("Running in dry-run (fake) mode")
	}

	store, err := models.NewSQLiteStore(cfg.Storage.Path)
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	bus := eventbus.New()

	// Set up Kubernetes client
	var kubeClient kube.Client
	switch cfg.Kubernetes.Mode {
	case "in-cluster":
		kubeClient, err = kube.NewInClusterClient()
	case "kubeconfig":
		kubeClient, err = kube.NewKubeconfigClient(cfg.Kubernetes.KubeconfigPath)
	default:
		log.Println("Using fake Kubernetes client")
		kubeClient = kube.NewFakeClient()
	}
	if err != nil {
		return fmt.Errorf("init kube client: %w", err)
	}

	// Create engines
	rcaEngine := rca.NewEngine(kubeClient, cfg.LLM, store)
	remEngine := remediation.NewEngine(kubeClient, cfg.Remediation.MaxRetries, cfg.Remediation.AutoApprove)

	// Start orchestrator
	orch := orchestrator.New(store, bus, rcaEngine, remEngine, kubeClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := orch.Run(ctx); err != nil {
			log.Printf("Orchestrator stopped: %v", err)
		}
	}()

	// Start detector (if enabled)
	if cfg.Detector.Enabled {
		det := detector.New(kubeClient, bus, cfg.Detector.CheckInterval, "")
		for _, rule := range rules.All() {
			det.RegisterRule(rule)
		}
		go func() {
			if err := det.Run(ctx); err != nil {
				log.Printf("Detector stopped: %v", err)
			}
		}()
		log.Println("Detector: started with 5 rules")
	}

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down...")
		cancel()
	}()

	server := api.NewServer(cfg.Server, store, bus)
	return server.Start()
}

func init() {
	_ = godotenv.Load()

	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Printf("Warning: could not create data dir: %v", err)
	}
}
