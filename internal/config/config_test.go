package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file with paths that exist
	configData := `
server:
  host: "0.0.0.0"
  port: 8080

logging:
  level: "info"
  format: "json"

metrics:
  collection:
    default_interval: "5m"

filesystems:
  - name: "root"
    mount_point: "/tmp"
    device: "sda1"
    interval: "1m"
  - name: "data"
    mount_point: "/var"
    device: "sdb1"

directories:
  general:
    path: "/tmp"
    subdirectory_levels: 0
    interval: "10m"
  minio:
    path: "/var"
    subdirectory_levels: 0
`

	tmpfile, err := os.CreateTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpfile.WriteString(configData); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading the config
	cfg, err := LoadConfig(tmpfile.Name(), false)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded config
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", cfg.Logging.Format)
	}

	if cfg.Metrics.Collection.DefaultInterval.Seconds() != 300 { // 5m = 300s
		t.Errorf("Expected default interval 300, got %d", cfg.Metrics.Collection.DefaultInterval.Seconds())
	}

	if len(cfg.Filesystems) != 2 {
		t.Errorf("Expected 2 filesystems, got %d", len(cfg.Filesystems))
	}

	if len(cfg.Directories) != 2 {
		t.Errorf("Expected 2 directories, got %d", len(cfg.Directories))
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Create a minimal config file
	configData := `
metrics:
  collection:
    default_interval: "5m"

filesystems:
  - name: "root"
    mount_point: "/tmp"
    device: "sda1"

directories:
  test:
    path: "/tmp"
    subdirectory_levels: 0
`

	tmpfile, err := os.CreateTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpfile.WriteString(configData); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading the config
	cfg, err := LoadConfig(tmpfile.Name(), false)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify defaults are set
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default level 'info', got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Expected default format 'json', got '%s'", cfg.Logging.Format)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	// Create a config file with invalid YAML
	configData := `
server:
  host: "0.0.0.0"
  port: 8080
  invalid: [unclosed bracket
`

	tmpfile, err := os.CreateTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpfile.WriteString(configData); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = Load(tmpfile.Name())
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name: "missing filesystems",
			config: `
metrics:
  collection:
    default_interval_seconds: 300
directories:
  test:
    path: "/tmp"
    subdirectory_levels: 0
`,
			wantErr: true,
		},
		{
			name: "missing directories",
			config: `
metrics:
  collection:
    default_interval_seconds: 300
filesystems:
  - name: "root"
    mount_point: "/tmp"
    device: "sda1"
`,
			wantErr: true,
		},
		{
			name: "invalid port",
			config: `
server:
  port: 70000
metrics:
  collection:
    default_interval_seconds: 300
filesystems:
  - name: "root"
    mount_point: "/tmp"
    device: "sda1"
directories:
  test:
    path: "/tmp"
    subdirectory_levels: 0
`,
			wantErr: true,
		},
		{
			name: "invalid logging level",
			config: `
logging:
  level: "invalid"
metrics:
  collection:
    default_interval_seconds: 300
filesystems:
  - name: "root"
    mount_point: "/tmp"
    device: "sda1"
directories:
  test:
    path: "/tmp"
    subdirectory_levels: 0
`,
			wantErr: true,
		},
		{
			name: "invalid interval",
			config: `
metrics:
  collection:
    default_interval_seconds: 0
filesystems:
  - name: "root"
    mount_point: "/tmp"
    device: "sda1"
directories:
  test:
    path: "/tmp"
    subdirectory_levels: 0
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "config_test")
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				if err := os.Remove(tmpfile.Name()); err != nil {
					t.Logf("Failed to remove temp file: %v", err)
				}
			}()

			if _, err := tmpfile.WriteString(tt.config); err != nil {
				t.Fatal(err)
			}

			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			_, err = Load(tmpfile.Name())
			if tt.wantErr && err == nil {
				t.Error("Expected validation error, got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDurationParsing(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Duration
		shouldErr bool
	}{
		{"seconds", "60s", 60 * time.Second, false},
		{"minutes", "30m", 30 * time.Minute, false},
		{"hours", "2h", 2 * time.Hour, false},
		{"mixed", "1h30m", 1*time.Hour + 30*time.Minute, false},
		{"complex", "1h30m45s", 1*time.Hour + 30*time.Minute + 45*time.Second, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration

			err := d.UnmarshalYAML(func(v interface{}) error {
				*(v.(*interface{})) = tt.input
				return nil
			})

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error for input %s, got none", tt.input)
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
				return
			}

			if d.Duration != tt.expected {
				t.Errorf("expected %v, got %v for input %s", tt.expected, d.Duration, tt.input)
			}
		})
	}
}

func TestDurationBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected time.Duration
	}{
		{"seconds", 60, 60 * time.Second},
		{"minutes", 1800, 1800 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration

			err := d.UnmarshalYAML(func(v interface{}) error {
				*(v.(*interface{})) = tt.input
				return nil
			})
			if err != nil {
				t.Errorf("unexpected error for input %d: %v", tt.input, err)
				return
			}

			if d.Duration != tt.expected {
				t.Errorf("expected %v, got %v for input %d", tt.expected, d.Duration, tt.input)
			}
		})
	}
}

func TestDurationSeconds(t *testing.T) {
	d := Duration{time.Duration(90) * time.Second}
	if d.Seconds() != 90 {
		t.Errorf("expected 90 seconds, got %d", d.Seconds())
	}
}
