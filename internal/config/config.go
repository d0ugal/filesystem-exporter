package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	promexporter_config "github.com/d0ugal/promexporter/config"
	"gopkg.in/yaml.v3"
)

// Use promexporter Duration type
type Duration = promexporter_config.Duration

type Config struct {
	promexporter_config.BaseConfig
	Filesystems []FilesystemConfig        `yaml:"filesystems"`
	Directories map[string]DirectoryGroup `yaml:"directories"`
}

type FilesystemConfig struct {
	Name       string   `yaml:"name"`
	MountPoint string   `yaml:"mount_point"`
	Device     string   `yaml:"device"`
	Interval   Duration `yaml:"interval"`
}

type DirectoryGroup struct {
	Path               string   `yaml:"path"`
	SubdirectoryLevels int      `yaml:"subdirectory_levels"`
	Interval           Duration `yaml:"interval"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	setDefaults(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}

	if !config.Metrics.Collection.DefaultIntervalSet {
		config.Metrics.Collection.DefaultInterval = promexporter_config.Duration{time.Second * 30}
	}

	if len(config.Filesystems) == 0 {
		config.Filesystems = []FilesystemConfig{
			{
				Name:       "root",
				MountPoint: "/",
				Device:     "root",
				Interval:   Duration{time.Minute * 5},
			},
		}
	}

	// Set defaults for filesystems
	for i := range config.Filesystems {
		if config.Filesystems[i].Interval.Duration == 0 {
			config.Filesystems[i].Interval = Duration{time.Minute * 5}
		}
	}

	// Set defaults for directories
	for name, group := range config.Directories {
		if group.Interval.Duration == 0 {
			group.Interval = Duration{time.Minute * 5}
			config.Directories[name] = group
		}
	}
}

// Validate performs comprehensive validation of the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if err := c.validateServerConfig(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	// Validate logging configuration
	if err := c.validateLoggingConfig(); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	// Validate metrics configuration
	if err := c.validateMetricsConfig(); err != nil {
		return fmt.Errorf("metrics config: %w", err)
	}

	// Validate filesystem configuration
	if err := c.validateFilesystemsConfig(); err != nil {
		return fmt.Errorf("filesystems config: %w", err)
	}

	// Validate directory configuration
	if err := c.validateDirectoriesConfig(); err != nil {
		return fmt.Errorf("directories config: %w", err)
	}

	return nil
}

func (c *Config) validateServerConfig() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Server.Port)
	}

	return nil
}

func (c *Config) validateLoggingConfig() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("invalid logging format: %s", c.Logging.Format)
	}

	return nil
}

func (c *Config) validateMetricsConfig() error {
	if c.Metrics.Collection.DefaultInterval.Seconds() < 1 {
		return fmt.Errorf("default interval must be at least 1 second, got %d", c.Metrics.Collection.DefaultInterval.Seconds())
	}

	if c.Metrics.Collection.DefaultInterval.Seconds() > 86400 {
		return fmt.Errorf("default interval must be at most 86400 seconds (24 hours), got %d", c.Metrics.Collection.DefaultInterval.Seconds())
	}

	return nil
}

func (c *Config) validateFilesystemsConfig() error {
	if len(c.Filesystems) == 0 {
		return fmt.Errorf("at least one filesystem must be specified")
	}

	for _, fs := range c.Filesystems {
		if fs.Name == "" {
			return fmt.Errorf("filesystem name cannot be empty")
		}
		if !filepath.IsAbs(fs.MountPoint) {
			return fmt.Errorf("filesystem mount point must be absolute: %s", fs.MountPoint)
		}
		if fs.Interval.Seconds() < 1 {
			return fmt.Errorf("filesystem interval must be at least 1 second, got %d", fs.Interval.Seconds())
		}
	}

	return nil
}

func (c *Config) validateDirectoriesConfig() error {
	for name, group := range c.Directories {
		if name == "" {
			return fmt.Errorf("directory group name cannot be empty")
		}
		if !filepath.IsAbs(group.Path) {
			return fmt.Errorf("directory path must be absolute: %s", group.Path)
		}
		if group.Interval.Seconds() < 1 {
			return fmt.Errorf("directory interval must be at least 1 second, got %d", group.Interval.Seconds())
		}
	}

	return nil
}

// GetDefaultInterval returns the default collection interval
func (c *Config) GetDefaultInterval() int {
	return c.Metrics.Collection.DefaultInterval.Seconds()
}

// GetFilesystemInterval returns the interval for a filesystem
func (c *Config) GetFilesystemInterval(fs FilesystemConfig) int {
	if fs.Interval.Duration > 0 {
		return fs.Interval.Seconds()
	}
	return c.GetDefaultInterval()
}

// GetDirectoryInterval returns the interval for a directory group
func (c *Config) GetDirectoryInterval(group DirectoryGroup) int {
	if group.Interval.Duration > 0 {
		return group.Interval.Seconds()
	}
	return c.GetDefaultInterval()
}
