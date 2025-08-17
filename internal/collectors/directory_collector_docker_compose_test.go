package collectors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
)

// TestSubdirectoryLevelsBehavior demonstrates the corrected behavior
// where setting subdirectory_levels: 2 collects directories at levels 0, 1, and 2
// (up to the specified level, inclusive)
func TestSubdirectoryLevelsBehavior(t *testing.T) {
	// Create a temporary directory structure that mimics the real setup
	tempDir := t.TempDir()

	// Create directory structure: /volume1/dougal/icloud-photos/2024/01/
	volume1Dir := filepath.Join(tempDir, "volume1")
	dougalDir := filepath.Join(volume1Dir, "dougal")
	icloudPhotosDir := filepath.Join(dougalDir, "icloud-photos")
	yearDir := filepath.Join(icloudPhotosDir, "2024")
	monthDir := filepath.Join(yearDir, "01")

	// Create the directories
	if err := os.MkdirAll(monthDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory structure: %v", err)
	}

	// Create test files at different levels
	testFiles := []string{
		filepath.Join(icloudPhotosDir, "photo1.jpg"), // Level 1
		filepath.Join(yearDir, "photo2.jpg"),         // Level 2
		filepath.Join(monthDir, "photo3.jpg"),        // Level 3
	}

	for _, file := range testFiles {
		content := strings.Repeat("test content ", 1000) // ~12KB per file
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test with subdirectory_levels: 2
	cfg := &config.Config{
		Metrics: config.MetricsConfig{
			Collection: config.CollectionConfig{
				DefaultInterval: config.Duration{Duration: 300 * time.Second},
			},
		},
		Directories: map[string]config.DirectoryGroup{
			"dougal": {
				Path:               dougalDir,
				SubdirectoryLevels: 2, // This should skip icloud-photos (level 1) and collect 2024 (level 2)
			},
		},
	}

	// Create metrics registry and collector
	registry := metrics.NewRegistry()
	collector := NewDirectoryCollector(cfg, registry)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start the collector
	collector.Start(ctx)

	// Give it time to collect metrics
	time.Sleep(2 * time.Second)

	// Test what gets collected with subdirectory_levels: 2 (should collect levels 0, 1, 2)
	t.Run("SubdirectoryLevels2", func(t *testing.T) {
		metrics := registry.GetRegistry()

		metricFamilies, err := metrics.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		expectedMetricName := "filesystem_exporter_directory_size_bytes"

		// With subdirectory_levels: 2, we should get:
		// - Base directory (dougalDir) at level 0
		// - icloud-photos at level 1
		// - yearDir (2024) at level 2
		expectedDirectories := []string{
			dougalDir,       // Base directory (level 0)
			icloudPhotosDir, // Level 1 directory
			yearDir,         // Level 2 directory (2024)
		}

		unexpectedDirectories := []string{
			// No unexpected directories - all levels up to 2 should be collected
		}

		// Check that expected directories are collected
		for _, expectedDir := range expectedDirectories {
			found := false

			for _, family := range metricFamilies {
				if family.GetName() == expectedMetricName {
					for _, metric := range family.GetMetric() {
						for _, label := range metric.GetLabel() {
							if label.GetName() == "directory" && label.GetValue() == expectedDir {
								found = true

								t.Logf("✓ Found expected metric for directory: %s", expectedDir)

								break
							}
						}

						if found {
							break
						}
					}
				}

				if found {
					break
				}
			}

			if !found {
				t.Errorf("❌ Expected directory %s was NOT collected", expectedDir)
			}
		}

		// Check that unexpected directories are NOT collected
		for _, unexpectedDir := range unexpectedDirectories {
			found := false

			for _, family := range metricFamilies {
				if family.GetName() == expectedMetricName {
					for _, metric := range family.GetMetric() {
						for _, label := range metric.GetLabel() {
							if label.GetName() == "directory" && label.GetValue() == unexpectedDir {
								found = true
								break
							}
						}

						if found {
							break
						}
					}
				}

				if found {
					break
				}
			}

			if found {
				t.Errorf("❌ Unexpected directory %s WAS collected (should be skipped)", unexpectedDir)
			} else {
				t.Logf("✓ Correctly skipped directory: %s", unexpectedDir)
			}
		}

		// Dump all collected metrics for verification
		t.Log("All collected directory metrics:")

		for _, family := range metricFamilies {
			if family.GetName() == expectedMetricName {
				for _, metric := range family.GetMetric() {
					labels := make([]string, 0, len(metric.GetLabel()))
					for _, label := range metric.GetLabel() {
						labels = append(labels, fmt.Sprintf("%s=%s", label.GetName(), label.GetValue()))
					}

					t.Logf("  %s{%s} = %f", family.GetName(), strings.Join(labels, ", "), metric.GetGauge().GetValue())
				}
			}
		}
	})
}
