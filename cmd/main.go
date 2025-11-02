package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"filesystem-exporter/internal/collectors"
	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/version"
	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/logging"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
)

var (
	// initOnce ensures main initialization only happens once
	// This prevents duplicate collector creation if main() is called multiple times
	initOnce sync.Once
	// collectorsCreated tracks if collectors have been created to detect duplication
	collectorsCreated bool
	collectorsMutex   sync.Mutex
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

	// Use sync.Once to ensure initialization only happens once per process
	// This prevents duplicate collector creation if main() is somehow called multiple times
	var filesystemCollector *collectors.FilesystemCollector
	var directoryCollector *collectors.DirectoryCollector
	var application *app.App
	var initErr error

	initOnce.Do(func() {
		slog.Info("Initializing filesystem-exporter (first time only)",
			"pid", os.Getpid(),
			"num_directories", len(cfg.Directories),
			"num_filesystems", len(cfg.Filesystems))

		// Check for duplicate initialization
		collectorsMutex.Lock()
		if collectorsCreated {
			slog.Error("CRITICAL: Collectors already created! Duplicate initialization detected",
				"pid", os.Getpid())
			collectorsMutex.Unlock()
			initErr = fmt.Errorf("collectors already created - duplicate initialization")
			return
		}
		collectorsCreated = true
		collectorsMutex.Unlock()

		// Initialize metrics registry using promexporter
		metricsRegistry := promexporter_metrics.NewRegistry("filesystem_exporter_info")

		// Add custom metrics to the registry
		filesystemRegistry := metrics.NewFilesystemRegistry(metricsRegistry)

		// Create collectors first (without app reference - will be set after build)
		// This allows us to register them with the builder before Build()
		slog.Info("Creating collectors",
			"num_directories", len(cfg.Directories),
			"num_filesystems", len(cfg.Filesystems))

		filesystemCollector = collectors.NewFilesystemCollector(cfg, filesystemRegistry, nil)
		directoryCollector = collectors.NewDirectoryCollector(cfg, filesystemRegistry, nil)

		slog.Info("Collectors created, registering with app builder",
			"filesystem_collector", fmt.Sprintf("%p", filesystemCollector),
			"directory_collector", fmt.Sprintf("%p", directoryCollector),
			"num_directories", len(cfg.Directories),
			"num_filesystems", len(cfg.Filesystems))

		// Build application with collectors registered BEFORE Build()
		application = app.New("Filesystem Exporter").
			WithConfig(&cfg.BaseConfig).
			WithMetrics(metricsRegistry).
			WithVersionInfo(version.Version, version.Commit, version.BuildDate).
			WithCollector(filesystemCollector).
			WithCollector(directoryCollector).
			Build()

		slog.Info("Application built, setting app references on collectors",
			"filesystem_collector", fmt.Sprintf("%p", filesystemCollector),
			"directory_collector", fmt.Sprintf("%p", directoryCollector))

		// Set app reference on collectors for tracing support (after build)
		// This is safe because Start() hasn't been called yet
		filesystemCollector.SetApp(application)
		directoryCollector.SetApp(application)
	})

	if initErr != nil {
		log.Fatalf("Failed to initialize: %v", initErr)
	}

	// Check if initialization was successful
	if filesystemCollector == nil || directoryCollector == nil || application == nil {
		slog.Error("CRITICAL: Initialization failed or was skipped - collectors are nil",
			"filesystem_collector_nil", filesystemCollector == nil,
			"directory_collector_nil", directoryCollector == nil,
			"application_nil", application == nil)
		log.Fatal("Failed to initialize collectors - initialization may have been called multiple times")
	}

	slog.Info("Initialization complete, starting application.Run()",
		"filesystem_collector", fmt.Sprintf("%p", filesystemCollector),
		"directory_collector", fmt.Sprintf("%p", directoryCollector),
		"pid", os.Getpid())

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
