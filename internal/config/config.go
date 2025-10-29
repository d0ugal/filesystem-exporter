package config

import (
	"embed"
	"fmt"
	"html/template"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	promexporter_config "github.com/d0ugal/promexporter/config"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.html
var templateFiles embed.FS

// Duration uses promexporter Duration type
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
func LoadConfig(path string, configFromEnv bool) (*Config, error) {
	if configFromEnv {
		return loadFromEnv()
	}

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

// loadFromEnv loads configuration from environment variables
func loadFromEnv() (*Config, error) {
	config := &Config{}

	// Load base configuration from environment
	baseConfig := &promexporter_config.BaseConfig{}

	// Server configuration
	if address := os.Getenv("FILESYSTEM_EXPORTER_SERVER_ADDRESS"); address != "" {
		// Parse host:port from address
		if host, portStr, err := net.SplitHostPort(address); err == nil {
			baseConfig.Server.Host = host

			if port, err := strconv.Atoi(portStr); err != nil {
				return nil, fmt.Errorf("invalid server port in address: %w", err)
			} else {
				baseConfig.Server.Port = port
			}
		} else {
			return nil, fmt.Errorf("invalid server address format: %w", err)
		}
	} else {
		baseConfig.Server.Host = "0.0.0.0"
		baseConfig.Server.Port = 8080
	}

	// Logging configuration
	if level := os.Getenv("FILESYSTEM_EXPORTER_LOGGING_LEVEL"); level != "" {
		baseConfig.Logging.Level = level
	} else {
		baseConfig.Logging.Level = "info"
	}

	if format := os.Getenv("FILESYSTEM_EXPORTER_LOGGING_FORMAT"); format != "" {
		baseConfig.Logging.Format = format
	} else {
		baseConfig.Logging.Format = "json"
	}

	// Metrics configuration
	if intervalStr := os.Getenv("FILESYSTEM_EXPORTER_METRICS_COLLECTION_DEFAULT_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err != nil {
			return nil, fmt.Errorf("invalid metrics default interval: %w", err)
		} else {
			baseConfig.Metrics.Collection.DefaultInterval = promexporter_config.Duration{Duration: interval}
			baseConfig.Metrics.Collection.DefaultIntervalSet = true
		}
	} else {
		baseConfig.Metrics.Collection.DefaultInterval = promexporter_config.Duration{Duration: time.Second * 30}
	}

	// Tracing configuration
	if enabledStr := os.Getenv("TRACING_ENABLED"); enabledStr != "" {
		enabled := enabledStr == "true"
		baseConfig.Tracing.Enabled = &enabled
	}

	if serviceName := os.Getenv("TRACING_SERVICE_NAME"); serviceName != "" {
		baseConfig.Tracing.ServiceName = serviceName
	}

	if endpoint := os.Getenv("TRACING_ENDPOINT"); endpoint != "" {
		baseConfig.Tracing.Endpoint = endpoint
	}

	config.BaseConfig = *baseConfig

	// Load directories from environment variables
	config.loadDirectoriesFromEnv()

	// Set defaults for any missing values
	setDefaults(config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// loadDirectoriesFromEnv loads directory configuration from environment variables
func (c *Config) loadDirectoriesFromEnv() {
	// Initialize the map if it's nil
	if c.Directories == nil {
		c.Directories = make(map[string]DirectoryGroup)
	}

	// Look for directory environment variables in the format FILESYSTEM_EXPORTER_DIRECTORIES_N_NAME, etc.
	for i := 0; i < 10; i++ { // Support up to 10 directories
		nameKey := fmt.Sprintf("FILESYSTEM_EXPORTER_DIRECTORIES_%d_NAME", i)
		pathKey := fmt.Sprintf("FILESYSTEM_EXPORTER_DIRECTORIES_%d_PATH", i)
		levelsKey := fmt.Sprintf("FILESYSTEM_EXPORTER_DIRECTORIES_%d_SUBDIRECTORY_LEVELS", i)

		name := os.Getenv(nameKey)
		if name == "" {
			continue // No more directories
		}

		path := os.Getenv(pathKey)
		if path == "" {
			continue // Path is required
		}

		levelsStr := os.Getenv(levelsKey)
		levels := 0

		if levelsStr != "" {
			if parsedLevels, err := strconv.Atoi(levelsStr); err == nil {
				levels = parsedLevels
			}
		}

		directory := DirectoryGroup{
			Path:               path,
			SubdirectoryLevels: levels,
		}

		c.Directories[name] = directory
		fmt.Printf("Loaded directory from env: name=%s, path=%s, levels=%d\n", name, path, levels)
	}

	fmt.Printf("Total directories loaded: %d\n", len(c.Directories))
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
		config.Metrics.Collection.DefaultInterval = promexporter_config.Duration{Duration: time.Second * 30}
	}

	if len(config.Filesystems) == 0 {
		config.Filesystems = []FilesystemConfig{
			{
				Name:       "root",
				MountPoint: "/",
				Device:     "root",
				Interval:   Duration{Duration: time.Minute * 5},
			},
		}
	}

	// Set defaults for filesystems
	for i := range config.Filesystems {
		if config.Filesystems[i].Interval.Duration == 0 {
			config.Filesystems[i].Interval = Duration{Duration: time.Minute * 5}
		}
	}

	// Set defaults for directories
	for name, group := range config.Directories {
		if group.Interval.Duration == 0 {
			group.Interval = Duration{Duration: time.Minute * 5}
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

// GetDisplayConfig returns configuration data safe for display
// Overrides BaseConfig to include filesystem and directory configuration
func (c *Config) GetDisplayConfig() map[string]interface{} {
	// Get base configuration
	config := c.BaseConfig.GetDisplayConfig()

	// Add filesystem configuration
	if len(c.Filesystems) > 0 {
		filesystems := make([]map[string]string, len(c.Filesystems))
		for i, fs := range c.Filesystems {
			filesystems[i] = map[string]string{
				"name":        fs.Name,
				"mount_point": fs.MountPoint,
				"device":      fs.Device,
				"interval":    fs.Interval.String(),
			}
		}

		config["Filesystems"] = filesystems
	} else {
		config["Filesystems"] = "None configured"
	}

	// Add directory configuration
	if len(c.Directories) > 0 {
		directories := make(map[string]map[string]interface{})
		for name, dir := range c.Directories {
			directories[name] = map[string]interface{}{
				"path":                dir.Path,
				"subdirectory_levels": dir.SubdirectoryLevels,
				"interval":            dir.Interval.String(),
			}
		}

		config["Directories"] = directories
	} else {
		config["Directories"] = "None configured"
	}

	return config
}

// RenderConfigHTML provides custom HTML fragments for specific configuration keys
func (c *Config) RenderConfigHTML(key string, value interface{}) (string, bool) {
	fmt.Printf("DEBUG: RenderConfigHTML called for key: %s\n", key)

	switch key {
	case "Directories":
		// Load and render the directories template
		html, ok := c.renderTemplate("directories", value)
		fmt.Printf("DEBUG: Directories template rendered: %s\n", html)

		return html, ok
	case "Filesystems":
		// Load and render the filesystems template
		html, ok := c.renderTemplate("filesystems", value)
		fmt.Printf("DEBUG: Filesystems template rendered: %s\n", html)

		return html, ok
	}

	fmt.Printf("DEBUG: No custom HTML for key: %s\n", key)

	return "", false
}

// renderTemplate loads and renders a template with the given data
func (c *Config) renderTemplate(templateName string, data interface{}) (string, bool) {
	tmpl, err := template.ParseFS(templateFiles, "templates/"+templateName+".html")
	if err != nil {
		// Fallback to hardcoded HTML if template file doesn't exist
		switch templateName {
		case "directories":
			return c.renderDirectoriesHTML(data), true
		case "filesystems":
			return c.renderFilesystemsHTML(data), true
		}

		return "", false
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", false
	}

	return buf.String(), true
}

// renderDirectoriesHTML renders the directories configuration as HTML
func (c *Config) renderDirectoriesHTML(data interface{}) string {
	directories, ok := data.(map[string]map[string]interface{})
	if !ok {
		return ""
	}

	html := `<div class="nested-map-container">`
	for name, config := range directories {
		html += `<div class="nested-map-item">`
		html += `<span class="nested-map-key">` + name + `:</span>`

		html += `<div class="object-container">`
		for k, v := range config {
			html += `<div class="object-item">`
			html += `<span class="object-key">` + k + `:</span>`
			html += `<span class="object-value">` + fmt.Sprintf("%v", v) + `</span>`
			html += `</div>`
		}

		html += `</div></div>`
	}

	html += `</div>`

	return html
}

// renderFilesystemsHTML renders the filesystems configuration as HTML
func (c *Config) renderFilesystemsHTML(data interface{}) string {
	filesystems, ok := data.([]map[string]string)
	if !ok {
		return ""
	}

	html := `<div class="array-container">`
	for i, item := range filesystems {
		html += `<div class="array-item">`
		html += `<span class="array-index">[` + fmt.Sprintf("%d", i) + `]</span>`

		html += `<div class="object-container">`
		for k, v := range item {
			html += `<div class="object-item">`
			html += `<span class="object-key">` + k + `:</span>`
			html += `<span class="object-value">` + v + `</span>`
			html += `</div>`
		}

		html += `</div></div>`
	}

	html += `</div>`

	return html
}

// GetLogging returns the logging configuration
func (c *Config) GetLogging() *promexporter_config.LoggingConfig {
	return c.BaseConfig.GetLogging()
}

// GetServer returns the server configuration
func (c *Config) GetServer() *promexporter_config.ServerConfig {
	return c.BaseConfig.GetServer()
}
