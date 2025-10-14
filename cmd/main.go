package main

import (
	"flag"
	"log/slog"
	"os"

	"filesystem-exporter/internal/collectors"
	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"

	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/logging"
	"github.com/d0ugal/promexporter/version"
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
		versionInfo := version.Get()
		slog.Info("filesystem-exporter version", "version", versionInfo.Version, "commit", versionInfo.Commit, "build_date", versionInfo.BuildDate, "go_version", versionInfo.GoVersion)
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

	// Load configuration
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

	// Initialize metrics
	metricsRegistry := metrics.NewFilesystemRegistry()

	// Set version info metric
	versionInfo := version.Get()
	metricsRegistry.VersionInfo.WithLabelValues(versionInfo.Version, versionInfo.Commit, versionInfo.BuildDate).Set(1)

	// Create collectors
	filesystemCollector := collectors.NewFilesystemCollector(cfg, metricsRegistry)
	directoryCollector := collectors.NewDirectoryCollector(cfg, metricsRegistry)

	// Build and run the application
	app.New("filesystem-exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(metricsRegistry.Registry).
		WithCollector(filesystemCollector).
		WithCollector(directoryCollector).
		Build().
		Run()
}
