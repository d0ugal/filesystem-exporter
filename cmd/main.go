package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/coordinator"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/version"
	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/logging"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
)

func main() {
	// Parse command line flags
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information")

	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	// Show version if requested
	if showVersion {
		slog.Info("filesystem-exporter version", "version", version.Version, "commit", version.Commit, "build_date", version.BuildDate)
		os.Exit(0)
	}

	if os.Getenv("FILESYSTEM_EXPORTER_CONFIG_FROM_ENV") == "true" {
		fmt.Fprintln(os.Stderr, "Warning: FILESYSTEM_EXPORTER_CONFIG_FROM_ENV is deprecated and has no effect. Env vars are always applied on top of yaml config.")
	}

	// Use environment variable if config flag is not provided
	if configPath == "" {
		if envConfig := os.Getenv("CONFIG_PATH"); envConfig != "" {
			configPath = envConfig
		} else {
			configPath = "config.yaml"
		}
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err, "path", configPath) //nolint:gosec // G706: configPath is from CLI/env, not HTTP input; slog structures the value safely
		os.Exit(1)
	}

	// Configure logging
	logging.Configure(&logging.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	// Validation already checks that at least one filesystem or directory is configured
	slog.Info("Initializing filesystem-exporter", //nolint:gosec // G706: values are from trusted config/system, not user HTTP input
		"pid", os.Getpid(),
		"num_directories", len(cfg.Directories),
		"num_filesystems", len(cfg.Filesystems))

	// Initialize metrics registry using promexporter
	metricsRegistry := promexporter_metrics.NewRegistry("filesystem_exporter_info")

	// Add custom metrics to the registry
	filesystemRegistry := metrics.NewFilesystemRegistry(metricsRegistry)

	// Build application with promexporter
	application := app.New("Filesystem Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(metricsRegistry).
		WithVersionInfo(version.Version, version.Commit, version.BuildDate).
		Build()

	// Get tracer for OTEL tracing
	tracer := application.GetTracer()

	// Create coordinator
	coord := coordinator.NewCoordinator(cfg, filesystemRegistry, tracer)

	// Start coordinator (will be started when app.Run() is called)
	// We need to pass a context - use context.Background() for now
	// The app will manage its own context
	coord.Start(context.Background())

	slog.Info("Initialization complete, starting application.Run()",
		"pid", os.Getpid())

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
