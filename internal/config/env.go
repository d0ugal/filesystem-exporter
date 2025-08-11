package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variable prefix
const EnvPrefix = "FILESYSTEM_EXPORTER_"

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}

	// Load server configuration
	if err := loadServerConfig(cfg); err != nil {
		return nil, fmt.Errorf("server config: %w", err)
	}

	// Load logging configuration
	if err := loadLoggingConfig(cfg); err != nil {
		return nil, fmt.Errorf("logging config: %w", err)
	}

	// Load metrics configuration
	if err := loadMetricsConfig(cfg); err != nil {
		return nil, fmt.Errorf("metrics config: %w", err)
	}

	// Load filesystems configuration
	if err := loadFilesystemsConfig(cfg); err != nil {
		return nil, fmt.Errorf("filesystems config: %w", err)
	}

	// Load directories configuration
	if err := loadDirectoriesConfig(cfg); err != nil {
		return nil, fmt.Errorf("directories config: %w", err)
	}

	// Set defaults
	setDefaults(cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func loadServerConfig(cfg *Config) error {
	if host := getEnv(EnvPrefix + "SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	if portStr := getEnv(EnvPrefix + "SERVER_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid server port: %s", portStr)
		}

		cfg.Server.Port = port
	}

	return nil
}

func loadLoggingConfig(cfg *Config) error {
	if level := getEnv(EnvPrefix + "LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}

	if format := getEnv(EnvPrefix + "LOG_FORMAT"); format != "" {
		cfg.Logging.Format = format
	}

	return nil
}

func loadMetricsConfig(cfg *Config) error {
	if intervalStr := getEnv(EnvPrefix + "METRICS_DEFAULT_INTERVAL"); intervalStr != "" {
		duration, err := time.ParseDuration(intervalStr)
		if err != nil {
			return fmt.Errorf("invalid metrics default interval: %s", intervalStr)
		}

		cfg.Metrics.Collection.DefaultInterval = Duration{duration}
		cfg.Metrics.Collection.DefaultIntervalSet = true
	}

	return nil
}

func loadFilesystemsConfig(cfg *Config) error {
	// Check if filesystems are configured via environment
	filesystemsStr := getEnv(EnvPrefix + "FILESYSTEMS")
	if filesystemsStr == "" {
		return nil // No filesystems configured via env
	}

	// Parse filesystems from comma-separated list
	// Format: "name1:mount1:device1:interval1,name2:mount2:device2:interval2"
	filesystemList := strings.Split(filesystemsStr, ",")

	for i, fsStr := range filesystemList {
		parts := strings.Split(strings.TrimSpace(fsStr), ":")
		if len(parts) < 3 || len(parts) > 4 {
			return fmt.Errorf("invalid filesystem format at index %d: expected 'name:mount:device[:interval]', got '%s'", i, fsStr)
		}

		fs := FilesystemConfig{
			Name:       strings.TrimSpace(parts[0]),
			MountPoint: strings.TrimSpace(parts[1]),
			Device:     strings.TrimSpace(parts[2]),
		}

		// Optional interval
		if len(parts) == 4 && parts[3] != "" {
			duration, err := time.ParseDuration(strings.TrimSpace(parts[3]))
			if err != nil {
				return fmt.Errorf("invalid interval for filesystem %s: %s", fs.Name, parts[3])
			}

			fs.Interval = &Duration{duration}
		}

		cfg.Filesystems = append(cfg.Filesystems, fs)
	}

	return nil
}

func loadDirectoriesConfig(cfg *Config) error {
	// Check if directories are configured via environment
	directoriesStr := getEnv(EnvPrefix + "DIRECTORIES")
	if directoriesStr == "" {
		return nil // No directories configured via env
	}

	// Parse directories from comma-separated list
	// Format: "name1:path1:levels1:interval1,name2:path2:levels2:interval2"
	directoryList := strings.Split(directoriesStr, ",")

	cfg.Directories = make(map[string]DirectoryGroup)

	for i, dirStr := range directoryList {
		parts := strings.Split(strings.TrimSpace(dirStr), ":")
		if len(parts) < 3 || len(parts) > 4 {
			return fmt.Errorf("invalid directory format at index %d: expected 'name:path:levels[:interval]', got '%s'", i, dirStr)
		}

		name := strings.TrimSpace(parts[0])
		path := strings.TrimSpace(parts[1])

		levels, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			return fmt.Errorf("invalid subdirectory levels for directory %s: %s", name, parts[2])
		}

		dir := DirectoryGroup{
			Path:               path,
			SubdirectoryLevels: levels,
		}

		// Optional interval
		if len(parts) == 4 && parts[3] != "" {
			duration, err := time.ParseDuration(strings.TrimSpace(parts[3]))
			if err != nil {
				return fmt.Errorf("invalid interval for directory %s: %s", name, parts[3])
			}

			dir.Interval = &Duration{duration}
		}

		cfg.Directories[name] = dir
	}

	return nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}

	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}

	if !cfg.Metrics.Collection.DefaultIntervalSet {
		cfg.Metrics.Collection.DefaultInterval = Duration{time.Second * 300}
	}
}

// getEnv gets an environment variable, returning empty string if not set
func getEnv(key string) string {
	return os.Getenv(key)
}

// LoadConfig loads configuration from either file or environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Check if environment-only mode is enabled
	if getEnv(EnvPrefix+"CONFIG_FROM_ENV") == "true" {
		return LoadFromEnv()
	}

	// If config file exists, load from file
	if _, err := os.Stat(configPath); err == nil {
		return Load(configPath)
	}

	// If no config file and environment variables are set, load from env
	if hasEnvConfig() {
		return LoadFromEnv()
	}

	// Fall back to file loading (will fail if file doesn't exist)
	return Load(configPath)
}

// hasEnvConfig checks if any configuration environment variables are set
func hasEnvConfig() bool {
	envVars := []string{
		EnvPrefix + "SERVER_HOST",
		EnvPrefix + "SERVER_PORT",
		EnvPrefix + "LOG_LEVEL",
		EnvPrefix + "LOG_FORMAT",
		EnvPrefix + "METRICS_DEFAULT_INTERVAL",
		EnvPrefix + "FILESYSTEMS",
		EnvPrefix + "DIRECTORIES",
	}

	for _, envVar := range envVars {
		if getEnv(envVar) != "" {
			return true
		}
	}

	return false
}
