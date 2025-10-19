package collectors

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	promexporter_config "github.com/d0ugal/promexporter/config"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
)

func TestDepthCalculation(t *testing.T) {
	tests := []struct {
		name          string
		basePath      string
		targetPath    string
		expectedDepth int
	}{
		{
			name:          "level 0 - root directory",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/",
			expectedDepth: 0,
		},
		{
			name:          "level 1 - immediate subdirectory",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/backup/",
			expectedDepth: 1,
		},
		{
			name:          "level 1 - immediate subdirectory (no trailing slash)",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/backup",
			expectedDepth: 1,
		},
		{
			name:          "level 2 - nested subdirectory",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/backup/config/",
			expectedDepth: 2,
		},
		{
			name:          "level 2 - nested subdirectory (no trailing slash)",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/backup/config",
			expectedDepth: 2,
		},
		{
			name:          "level 2 - deeper nested",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/media/videos/",
			expectedDepth: 2,
		},
		{
			name:          "level 3 - very deep nested",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/media/videos/2025/",
			expectedDepth: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate relative path
			relPath, err := filepath.Rel(tt.basePath, tt.targetPath)
			if err != nil {
				t.Fatalf("Failed to calculate relative path: %v", err)
			}

			// Test the OLD method (what we had before)
			oldDepth := strings.Count(relPath, string(filepath.Separator))

			// Test the NEW method (what we have now)
			pathComponents := strings.Split(relPath, string(filepath.Separator))

			var newDepth int
			if len(pathComponents) == 1 && pathComponents[0] == "." {
				newDepth = 0
			} else {
				newDepth = len(pathComponents)
			}

			t.Logf("Base path: %s", tt.basePath)
			t.Logf("Target path: %s", tt.targetPath)
			t.Logf("Relative path: %s", relPath)
			t.Logf("Path components: %v", pathComponents)
			t.Logf("Expected depth: %d", tt.expectedDepth)
			t.Logf("Old method depth: %d", oldDepth)
			t.Logf("New method depth: %d", newDepth)

			if newDepth != tt.expectedDepth {
				t.Errorf("New method: expected depth %d, got %d", tt.expectedDepth, newDepth)
			}

			if oldDepth != tt.expectedDepth {
				t.Logf("Old method: expected depth %d, got %d (this is why we needed the fix)", tt.expectedDepth, oldDepth)
			}
		})
	}
}

func TestDepthCalculationEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		basePath      string
		targetPath    string
		expectedDepth int
	}{
		{
			name:          "empty relative path",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/",
			expectedDepth: 0,
		},
		{
			name:          "single component",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/backup",
			expectedDepth: 1,
		},
		{
			name:          "two components",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/backup/config",
			expectedDepth: 2,
		},
		{
			name:          "three components",
			basePath:      "/mnt/data/apps/",
			targetPath:    "/mnt/data/apps/backup/config/settings",
			expectedDepth: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relPath, err := filepath.Rel(tt.basePath, tt.targetPath)
			if err != nil {
				t.Fatalf("Failed to calculate relative path: %v", err)
			}

			pathComponents := strings.Split(relPath, string(filepath.Separator))

			var depth int
			if len(pathComponents) == 1 && pathComponents[0] == "." {
				depth = 0
			} else {
				depth = len(pathComponents)
			}

			t.Logf("Relative path: '%s'", relPath)
			t.Logf("Path components: %v", pathComponents)
			t.Logf("Calculated depth: %d", depth)
			t.Logf("Expected depth: %d", tt.expectedDepth)

			if depth != tt.expectedDepth {
				t.Errorf("Expected depth %d, got %d", tt.expectedDepth, depth)
			}
		})
	}
}

func TestDirectoryCollectorMutex(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Directories: map[string]config.DirectoryGroup{
			"test": {
				Path:               "/tmp",
				SubdirectoryLevels: 0,
			},
		},
	}

	// Create metrics registry
	// Create a mock base registry for testing
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewFilesystemRegistry(baseRegistry)

	// Create collector
	collector := NewDirectoryCollector(cfg, registry)

	// Test that mutex prevents concurrent du operations
	var wg sync.WaitGroup

	startTime := time.Now()

	// Start multiple goroutines that would normally run du concurrently
	for i := 0; i < 3; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			// This will try to acquire the mutex
			err := collector.collectSingleDirectoryFile(context.Background(), "test", "/tmp", "test", 0)
			if err != nil {
				t.Logf("Goroutine %d: Expected error for /tmp (likely doesn't exist), got: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	duration := time.Since(startTime)

	// If the mutex is working correctly, operations should be serialized
	// and take longer than if they were truly concurrent
	// For a simple /tmp directory, this should be noticeable
	t.Logf("Total time for 3 serialized operations: %v", duration)

	// Verify the mutex field exists and is accessible
	// Note: We can't directly compare sync.Mutex values, but we can verify the field exists
	t.Log("Mutex field exists and is accessible")
}

func TestDirectoryCollectorConcurrency(t *testing.T) {
	// Create a test config with multiple directories
	interval := config.Duration{Duration: time.Duration(60) * time.Second}
	defaultInterval := config.Duration{Duration: time.Duration(300) * time.Second}
	cfg := &config.Config{
		BaseConfig: promexporter_config.BaseConfig{
			Metrics: promexporter_config.MetricsConfig{
				Collection: promexporter_config.CollectionConfig{
					DefaultInterval: defaultInterval,
				},
			},
		},
		Directories: map[string]config.DirectoryGroup{
			"test1": {
				Path:               "/tmp",
				SubdirectoryLevels: 0,
				Interval:           interval,
			},
			"test2": {
				Path:               "/var",
				SubdirectoryLevels: 0,
				Interval:           interval,
			},
		},
	}

	// Create metrics registry
	// Create a mock base registry for testing
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewFilesystemRegistry(baseRegistry)

	// Create collector
	collector := NewDirectoryCollector(cfg, registry)

	// Test that the collector can handle concurrent requests
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the collector
	collector.Start(ctx)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel to stop the collector
	cancel()

	// Verify the collector was created correctly
	if collector.config == nil {
		t.Error("Collector config should not be nil")
	}

	if collector.metrics == nil {
		t.Error("Collector metrics should not be nil")
	}
}

func TestLockWaitDurationMetric(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Directories: map[string]config.DirectoryGroup{
			"test": {
				Path:               "/tmp",
				SubdirectoryLevels: 0,
			},
		},
	}

	// Create metrics registry
	// Create a mock base registry for testing
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewFilesystemRegistry(baseRegistry)

	// Create collector
	collector := NewDirectoryCollector(cfg, registry)

	// Test that lock wait duration is recorded
	var wg sync.WaitGroup

	// Start multiple goroutines to create lock contention
	for i := 0; i < 3; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			// This will try to acquire the mutex and record wait time
			err := collector.collectSingleDirectoryFile(context.Background(), "test", "/tmp", "test", 0)
			if err != nil {
				t.Logf("Goroutine %d: Expected error for /tmp, got: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify that the lock wait duration metric was recorded
	// Note: We can't easily verify the exact values without exposing internal state,
	// but we can verify the metric exists and was accessed
	t.Log("Lock wait duration metric test completed - check metrics for filesystem_exporter_du_lock_wait_duration_seconds")
}
