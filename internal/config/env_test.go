package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnvBasic(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "basic server configuration",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_SERVER_HOST": "127.0.0.1",
				"FILESYSTEM_EXPORTER_SERVER_PORT": "9090",
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1",
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1",
			},
			validate: func(cfg *Config) error {
				if cfg.Server.Host != "127.0.0.1" {
					t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
				}
				if cfg.Server.Port != 9090 {
					t.Errorf("expected port 9090, got %d", cfg.Server.Port)
				}
				return nil
			},
		},
		{
			name: "logging configuration",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_LOG_LEVEL":   "debug",
				"FILESYSTEM_EXPORTER_LOG_FORMAT":  "text",
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1",
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1",
			},
			validate: func(cfg *Config) error {
				if cfg.Logging.Level != "debug" {
					t.Errorf("expected log level debug, got %s", cfg.Logging.Level)
				}
				if cfg.Logging.Format != "text" {
					t.Errorf("expected log format text, got %s", cfg.Logging.Format)
				}
				return nil
			},
		},
		{
			name: "metrics configuration",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_METRICS_DEFAULT_INTERVAL": "1m",
				"FILESYSTEM_EXPORTER_FILESYSTEMS":              "root:/:sda1",
				"FILESYSTEM_EXPORTER_DIRECTORIES":              "home:/home:1",
			},
			validate: func(cfg *Config) error {
				if cfg.Metrics.Collection.DefaultInterval.Duration != time.Minute {
					t.Errorf("expected interval 1m, got %v", cfg.Metrics.Collection.DefaultInterval.Duration)
				}
				return nil
			},
		},
	}

	runEnvTests(t, tests)
}

func TestLoadFromEnvFilesystems(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "filesystems configuration",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1:1m,data:/tmp:sdb1:2m",
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1",
			},
			validate: func(cfg *Config) error {
				if len(cfg.Filesystems) != 2 {
					t.Errorf("expected 2 filesystems, got %d", len(cfg.Filesystems))
				}

				// Check first filesystem
				if cfg.Filesystems[0].Name != "root" {
					t.Errorf("expected name root, got %s", cfg.Filesystems[0].Name)
				}
				if cfg.Filesystems[0].MountPoint != "/" {
					t.Errorf("expected mount point /, got %s", cfg.Filesystems[0].MountPoint)
				}
				if cfg.Filesystems[0].Device != "sda1" {
					t.Errorf("expected device sda1, got %s", cfg.Filesystems[0].Device)
				}
				if cfg.Filesystems[0].Interval == nil || cfg.Filesystems[0].Interval.Duration != time.Minute {
					t.Errorf("expected interval 1m, got %v", cfg.Filesystems[0].Interval)
				}

				// Check second filesystem
				if cfg.Filesystems[1].Name != "data" {
					t.Errorf("expected name data, got %s", cfg.Filesystems[1].Name)
				}
				if cfg.Filesystems[1].MountPoint != "/tmp" {
					t.Errorf("expected mount point /tmp, got %s", cfg.Filesystems[1].MountPoint)
				}
				if cfg.Filesystems[1].Device != "sdb1" {
					t.Errorf("expected device sdb1, got %s", cfg.Filesystems[1].Device)
				}
				if cfg.Filesystems[1].Interval == nil || cfg.Filesystems[1].Interval.Duration != 2*time.Minute {
					t.Errorf("expected interval 2m, got %v", cfg.Filesystems[1].Interval)
				}
				return nil
			},
		},
		{
			name: "filesystems without intervals",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1,data:/tmp:sdb1",
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1",
			},
			validate: func(cfg *Config) error {
				if len(cfg.Filesystems) != 2 {
					t.Errorf("expected 2 filesystems, got %d", len(cfg.Filesystems))
				}

				if cfg.Filesystems[0].Interval != nil {
					t.Errorf("expected no interval for first filesystem, got %v", cfg.Filesystems[0].Interval)
				}
				if cfg.Filesystems[1].Interval != nil {
					t.Errorf("expected no interval for second filesystem, got %v", cfg.Filesystems[1].Interval)
				}
				return nil
			},
		},
	}

	runEnvTests(t, tests)
}

func TestLoadFromEnvDirectories(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "directories configuration",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1:10m,logs:/tmp:0:5m",
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1",
			},
			validate: func(cfg *Config) error {
				if len(cfg.Directories) != 2 {
					t.Errorf("expected 2 directories, got %d", len(cfg.Directories))
				}

				// Check home directory
				home, exists := cfg.Directories["home"]
				if !exists {
					t.Error("expected home directory to exist")
				}
				if home.Path != "/home" {
					t.Errorf("expected path /home, got %s", home.Path)
				}
				if home.SubdirectoryLevels != 1 {
					t.Errorf("expected subdirectory levels 1, got %d", home.SubdirectoryLevels)
				}
				if home.Interval == nil || home.Interval.Duration != 10*time.Minute {
					t.Errorf("expected interval 10m, got %v", home.Interval)
				}

				// Check logs directory
				logs, exists := cfg.Directories["logs"]
				if !exists {
					t.Error("expected logs directory to exist")
				}
				if logs.Path != "/tmp" {
					t.Errorf("expected path /tmp, got %s", logs.Path)
				}
				if logs.SubdirectoryLevels != 0 {
					t.Errorf("expected subdirectory levels 0, got %d", logs.SubdirectoryLevels)
				}
				if logs.Interval == nil || logs.Interval.Duration != 5*time.Minute {
					t.Errorf("expected interval 5m, got %v", logs.Interval)
				}
				return nil
			},
		},
		{
			name: "directories without intervals",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1,logs:/tmp:0",
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1",
			},
			validate: func(cfg *Config) error {
				if len(cfg.Directories) != 2 {
					t.Errorf("expected 2 directories, got %d", len(cfg.Directories))
				}

				home := cfg.Directories["home"]
				if home.Interval != nil {
					t.Errorf("expected no interval for home directory, got %v", home.Interval)
				}

				logs := cfg.Directories["logs"]
				if logs.Interval != nil {
					t.Errorf("expected no interval for logs directory, got %v", logs.Interval)
				}
				return nil
			},
		},
	}

	runEnvTests(t, tests)
}

func TestLoadFromEnvDirectoriesWithColons(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "directories with paths containing colons",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_DIRECTORIES": "hoose:/usr/share/hoose/:0,frigate:/mnt/media/frigate/:1",
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1",
			},
			expectError: false, // Let's see what actually gets parsed
			validate: func(cfg *Config) error {
				t.Logf("Parsed directories: %+v", cfg.Directories)
				for name, dir := range cfg.Directories {
					t.Logf("Directory %s: path=%s, levels=%d", name, dir.Path, dir.SubdirectoryLevels)
				}
				return nil
			},
		},
	}

	runEnvTests(t, tests)
}

func TestLoadFromEnvErrors(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "invalid server port",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_SERVER_PORT": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid metrics interval",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_METRICS_DEFAULT_INTERVAL": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid filesystem format",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1:1m,invalid",
			},
			expectError: true,
		},
		{
			name: "invalid filesystem interval",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1:invalid",
			},
			expectError: true,
		},
		{
			name: "invalid directory format",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1:10m,invalid",
			},
			expectError: true,
		},
		{
			name: "invalid directory levels",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:invalid:10m",
			},
			expectError: true,
		},
		{
			name: "invalid directory interval",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1:invalid",
			},
			expectError: true,
		},
	}

	runEnvTests(t, tests)
}

func runEnvTests(t *testing.T, tests []struct {
	name        string
	envVars     map[string]string
	expectError bool
	validate    func(*Config) error
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("failed to set environment variable %s: %v", key, err)
				}
			}

			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					if err := os.Unsetenv(key); err != nil {
						t.Logf("failed to unset environment variable %s: %v", key, err)
					}
				}
			}()

			// Load configuration
			cfg, err := LoadFromEnv()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				if err := tt.validate(cfg); err != nil {
					t.Errorf("validation failed: %v", err)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		configPath  string
		expectError bool
	}{
		{
			name: "force environment mode",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_CONFIG_FROM_ENV": "true",
				"FILESYSTEM_EXPORTER_SERVER_PORT":     "9090",
				"FILESYSTEM_EXPORTER_FILESYSTEMS":     "root:/:sda1",
				"FILESYSTEM_EXPORTER_DIRECTORIES":     "home:/home:1",
			},
			configPath: "nonexistent.yaml",
		},
		{
			name: "environment variables present but no force flag",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_SERVER_PORT": "9090",
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1",
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1",
			},
			configPath: "nonexistent.yaml",
		},
		{
			name:        "no environment variables, no config file",
			configPath:  "nonexistent.yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("failed to set environment variable %s: %v", key, err)
				}
			}

			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					if err := os.Unsetenv(key); err != nil {
						t.Logf("failed to unset environment variable %s: %v", key, err)
					}
				}
			}()

			// Load configuration
			cfg, err := LoadConfig(tt.configPath)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Basic validation that config was loaded
			if cfg == nil {
				t.Error("expected config, got nil")
			}
		})
	}
}

func TestHasEnvConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no environment variables",
			expected: false,
		},
		{
			name: "server host set",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_SERVER_HOST": "127.0.0.1",
			},
			expected: true,
		},
		{
			name: "server port set",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_SERVER_PORT": "9090",
			},
			expected: true,
		},
		{
			name: "log level set",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_LOG_LEVEL": "debug",
			},
			expected: true,
		},
		{
			name: "log format set",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_LOG_FORMAT": "text",
			},
			expected: true,
		},
		{
			name: "metrics interval set",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_METRICS_DEFAULT_INTERVAL": "1m",
			},
			expected: true,
		},
		{
			name: "filesystems set",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_FILESYSTEMS": "root:/:sda1",
			},
			expected: true,
		},
		{
			name: "directories set",
			envVars: map[string]string{
				"FILESYSTEM_EXPORTER_DIRECTORIES": "home:/home:1",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("failed to set environment variable %s: %v", key, err)
				}
			}

			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					if err := os.Unsetenv(key); err != nil {
						t.Logf("failed to unset environment variable %s: %v", key, err)
					}
				}
			}()

			result := hasEnvConfig()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
