package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"filesystem-exporter/internal/collectors"
	"filesystem-exporter/internal/config"
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

	// Initialize metrics registry using promexporter
	metricsRegistry := promexporter_metrics.NewRegistry("filesystem_exporter_info")

	// Set version info metric with filesystem-exporter version information
	metricsRegistry.VersionInfo.WithLabelValues(version.Version, version.Commit, version.BuildDate).Set(1)

	// Add custom metrics to the registry
	filesystemRegistry := metrics.NewFilesystemRegistry(metricsRegistry)

	// Create collectors
	filesystemCollector := collectors.NewFilesystemCollector(cfg, filesystemRegistry)
	directoryCollector := collectors.NewDirectoryCollector(cfg, filesystemRegistry)

	// Build and run the application
	if err := app.New("Filesystem Exporter").
		WithConfig(cfg).
		WithMetrics(metricsRegistry).
		WithCollector(filesystemCollector).
		WithCollector(directoryCollector).
		Build().
		Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
