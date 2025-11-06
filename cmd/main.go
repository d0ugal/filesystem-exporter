package main

import (
	"context"
	"flag"
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

	// Check if we should use environment variables
	configFromEnv := os.Getenv("FILESYSTEM_EXPORTER_CONFIG_FROM_ENV") == "true"

	// Load configuration
	var (
		cfg *config.Config
		err error
	)

	if configFromEnv {
		cfg, err = config.LoadConfig("", true)
	} else {
		// Use environment variable if config flag is not provided
		if configPath == "" {
			if envConfig := os.Getenv("CONFIG_PATH"); envConfig != "" {
				configPath = envConfig
			} else {
				configPath = "config.yaml"
			}
		}

		cfg, err = config.LoadConfig(configPath, false)
	}

	if err != nil {
		slog.Error("Failed to load configuration", "error", err, "path", configPath)
		os.Exit(1)
	}

	// Configure logging
	logging.Configure(&logging.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	// Validation already checks that at least one filesystem or directory is configured
	slog.Info("Initializing filesystem-exporter",
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
