package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration represents a time duration that can be parsed from strings
type Duration struct {
	time.Duration
}

// UnmarshalYAML implements custom unmarshaling for duration strings
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value interface{}
	if err := unmarshal(&value); err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		duration, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid duration format '%s': %w", v, err)
		}

		d.Duration = duration
	case int:
		// Backward compatibility: treat as seconds
		d.Duration = time.Duration(v) * time.Second
	case int64:
		// Backward compatibility: treat as seconds
		d.Duration = time.Duration(v) * time.Second
	default:
		return fmt.Errorf("duration must be a string (e.g., '60s', '1h') or integer (seconds)")
	}

	return nil
}

// Seconds returns the duration in seconds
func (d *Duration) Seconds() int {
	return int(d.Duration.Seconds())
}

type Config struct {
	Server      ServerConfig              `yaml:"server"`
	Logging     LoggingConfig             `yaml:"logging"`
	Metrics     MetricsConfig             `yaml:"metrics"`
	Filesystems []FilesystemConfig        `yaml:"filesystems"`
	Directories map[string]DirectoryGroup `yaml:"directories"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // "json" or "text"
}

type MetricsConfig struct {
	Collection CollectionConfig `yaml:"collection"`
}

type CollectionConfig struct {
	DefaultInterval Duration `yaml:"default_interval"`
	// Track if the value was explicitly set
	DefaultIntervalSet bool `yaml:"-"`
}

// UnmarshalYAML implements custom unmarshaling to track if the value was set
func (c *CollectionConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Create a temporary struct to unmarshal into
	type tempCollectionConfig struct {
		DefaultInterval Duration `yaml:"default_interval"`
	}

	var temp tempCollectionConfig
	if err := unmarshal(&temp); err != nil {
		return err
	}

	c.DefaultInterval = temp.DefaultInterval
	c.DefaultIntervalSet = true

	return nil
}

type FilesystemConfig struct {
	Name       string    `yaml:"name"`
	MountPoint string    `yaml:"mount_point"`
	Device     string    `yaml:"device"`
	Interval   *Duration `yaml:"interval,omitempty"`
}

type DirectoryGroup struct {
	Path               string    `yaml:"path"`
	SubdirectoryLevels int       `yaml:"subdirectory_levels"`
	Interval           *Duration `yaml:"interval,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
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
		config.Metrics.Collection.DefaultInterval = Duration{time.Second * 300}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
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

	// Validate filesystems
	if err := c.validateFilesystems(); err != nil {
		return fmt.Errorf("filesystems config: %w", err)
	}

	// Validate directories
	if err := c.validateDirectories(); err != nil {
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

func (c *Config) validateFilesystems() error {
	if len(c.Filesystems) == 0 {
		return fmt.Errorf("at least one filesystem must be configured")
	}

	seenNames := make(map[string]bool)
	seenMountPoints := make(map[string]bool)

	for i, fs := range c.Filesystems {
		// Validate name
		if fs.Name == "" {
			return fmt.Errorf("filesystem %d: name is required", i)
		}

		if seenNames[fs.Name] {
			return fmt.Errorf("filesystem %d: duplicate name '%s'", i, fs.Name)
		}

		seenNames[fs.Name] = true

		// Validate mount point
		if fs.MountPoint == "" {
			return fmt.Errorf("filesystem %d: mount_point is required", i)
		}

		if !filepath.IsAbs(fs.MountPoint) {
			return fmt.Errorf("filesystem %d: mount_point must be absolute path, got '%s'", i, fs.MountPoint)
		}

		if seenMountPoints[fs.MountPoint] {
			return fmt.Errorf("filesystem %d: duplicate mount_point '%s'", i, fs.MountPoint)
		}

		seenMountPoints[fs.MountPoint] = true

		// Validate device
		if fs.Device == "" {
			return fmt.Errorf("filesystem %d: device is required", i)
		}

		// Validate interval
		if fs.Interval != nil {
			if fs.Interval.Seconds() < 1 {
				return fmt.Errorf("filesystem %d: interval must be at least 1 second, got %d", i, fs.Interval.Seconds())
			}

			if fs.Interval.Seconds() > 86400 {
				return fmt.Errorf("filesystem %d: interval must be at most 86400 seconds (24 hours), got %d", i, fs.Interval.Seconds())
			}
		}

		// Validate mount point exists (if possible)
		if _, err := os.Stat(fs.MountPoint); err != nil {
			return fmt.Errorf("filesystem %d: mount_point '%s' does not exist or is not accessible", i, fs.MountPoint)
		}
	}

	return nil
}

func (c *Config) validateDirectories() error {
	if len(c.Directories) == 0 {
		return fmt.Errorf("at least one directory must be configured")
	}

	seenPaths := make(map[string]bool)

	for name, dir := range c.Directories {
		// Validate name
		if name == "" {
			return fmt.Errorf("directory: name is required")
		}

		// Validate path
		if dir.Path == "" {
			return fmt.Errorf("directory '%s': path is required", name)
		}

		if !filepath.IsAbs(dir.Path) {
			return fmt.Errorf("directory '%s': path must be absolute, got '%s'", name, dir.Path)
		}

		if seenPaths[dir.Path] {
			return fmt.Errorf("directory '%s': duplicate path '%s'", name, dir.Path)
		}

		seenPaths[dir.Path] = true

		// Validate subdirectory levels
		if dir.SubdirectoryLevels < 0 {
			return fmt.Errorf("directory '%s': subdirectory_levels must be non-negative, got %d", name, dir.SubdirectoryLevels)
		}

		if dir.SubdirectoryLevels > 10 {
			return fmt.Errorf("directory '%s': subdirectory_levels must be at most 10, got %d", name, dir.SubdirectoryLevels)
		}

		// Validate interval
		if dir.Interval != nil {
			if dir.Interval.Seconds() < 1 {
				return fmt.Errorf("directory '%s': interval must be at least 1 second, got %d", name, dir.Interval.Seconds())
			}

			if dir.Interval.Seconds() > 86400 {
				return fmt.Errorf("directory '%s': interval must be at most 86400 seconds (24 hours), got %d", name, dir.Interval.Seconds())
			}
		}

		// Validate path exists (if possible)
		if _, err := os.Stat(dir.Path); err != nil {
			return fmt.Errorf("directory '%s': path '%s' does not exist or is not accessible", name, dir.Path)
		}
	}

	return nil
}

// GetFilesystemInterval returns the effective interval for a filesystem
func (c *Config) GetFilesystemInterval(fs FilesystemConfig) int {
	if fs.Interval != nil {
		return fs.Interval.Seconds()
	}

	return c.Metrics.Collection.DefaultInterval.Seconds()
}

// GetDirectoryInterval returns the effective interval for a directory
func (c *Config) GetDirectoryInterval(dir DirectoryGroup) int {
	if dir.Interval != nil {
		return dir.Interval.Seconds()
	}

	return c.Metrics.Collection.DefaultInterval.Seconds()
}
