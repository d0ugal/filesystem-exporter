package main

import (
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

	// Build the app first (without the collector) so we can get its tracer
	// to wire into the coordinator. Then re-build with the collector attached
	// so app.Run() owns the coordinator's lifecycle and SIGTERM cancels its
	// context cleanly instead of leaking goroutines under context.Background().
	application := app.New("Filesystem Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(metricsRegistry).
		WithVersionInfo(version.Version, version.Commit, version.BuildDate).
		Build()

	tracer := application.GetTracer()
	coord := coordinator.NewCoordinator(cfg, filesystemRegistry, tracer)
	application.WithCollector(coord)

	slog.Info("Initialization complete, starting application.Run()",
		"pid", os.Getpid())

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
