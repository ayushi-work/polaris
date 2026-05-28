package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/ayushi/polaris/internal/api"
	"github.com/ayushi/polaris/internal/config"
	"github.com/ayushi/polaris/internal/eventbus"
	"github.com/ayushi/polaris/internal/models"
)

var (
	cfgPath  string
	dryRun   bool
	cfg      *config.Config
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
	var err error
	cfg, err = config.Load(cfgPath)
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

	server := api.NewServer(cfg.Server, store, bus)
	return server.Start()
}

func init() {
	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Printf("Warning: could not create data dir: %v", err)
	}
}
