package metrics

import (
	"testing"

	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestFilesystemRegistry(t *testing.T) {
	// Create a base registry
	baseRegistry := promexporter_metrics.NewRegistry("test")

	// Create filesystem registry
	registry := NewFilesystemRegistry(baseRegistry)

	// Use the metrics to ensure they're registered
	registry.VolumeSizeGauge.With(prometheus.Labels{"device": "test", "mount_point": "/", "volume": "test"}).Set(1)
	registry.VolumeAvailableGauge.With(prometheus.Labels{"device": "test", "mount_point": "/", "volume": "test"}).Set(1)
	registry.VolumeUsedRatioGauge.With(prometheus.Labels{"device": "test", "mount_point": "/", "volume": "test"}).Set(1)
	registry.DirectorySizeGauge.With(prometheus.Labels{"group": "test", "directory": "/", "mode": "test", "subdirectory_level": "0"}).Set(1)
	registry.CollectionDuration.With(prometheus.Labels{"group": "test", "interval_seconds": "60", "type": "test"}).Set(1)
	registry.CollectionSuccess.With(prometheus.Labels{"group": "test", "interval_seconds": "60", "type": "test"}).Inc()
	registry.CollectionFailedCounter.With(prometheus.Labels{"group": "test", "interval_seconds": "60", "type": "test"}).Inc()

	// Test that all documented metrics are registered
	expectedMetrics := []string{
		"filesystem_exporter_volume_size_bytes",
		"filesystem_exporter_volume_available_bytes",
		"filesystem_exporter_volume_used_ratio",
		"filesystem_exporter_directory_size_bytes",
		"filesystem_exporter_collection_duration_seconds",
		"filesystem_exporter_collection_success_total",
		"filesystem_exporter_collection_failed_total",
	}

	// Test that documented metrics exist
	for _, metricName := range expectedMetrics {
		t.Run("documented_metric_"+metricName, func(t *testing.T) {
			metricFamilies, err := testutil.GatherAndCount(baseRegistry.GetRegistry(), metricName)
			if err != nil {
				t.Fatalf("Failed to gather metrics: %v", err)
			}

			if metricFamilies == 0 {
				t.Errorf("Documented metric %s not found in registry", metricName)
			}
		})
	}

	// Use operational metrics to ensure they're registered
	registry.CollectionIntervalGauge.With(prometheus.Labels{"group": "test", "type": "test"}).Set(60)
	registry.CollectionTimestampGauge.With(prometheus.Labels{"group": "test", "interval_seconds": "60", "type": "test"}).Set(1)
	registry.DirectoriesFailedCounter.With(prometheus.Labels{"group": "test", "reason": "test"}).Inc()
	registry.DuLockWaitDurationGauge.With(prometheus.Labels{"group": "test", "path": "/test"}).Set(1)
	registry.DirectoriesProcessedCounter.With(prometheus.Labels{"group": "test", "method": "test"}).Inc()

	// Test that operational metrics exist (used by collectors)
	operationalMetrics := []string{
		"filesystem_exporter_collection_interval_seconds",
		"filesystem_exporter_collection_timestamp",
		"filesystem_directories_failed_total",
		"filesystem_exporter_du_lock_wait_duration_seconds",
		"filesystem_exporter_directories_processed_total",
	}

	for _, metricName := range operationalMetrics {
		t.Run("operational_metric_"+metricName, func(t *testing.T) {
			metricFamilies, err := testutil.GatherAndCount(baseRegistry.GetRegistry(), metricName)
			if err != nil {
				t.Fatalf("Failed to gather metrics: %v", err)
			}

			if metricFamilies == 0 {
				t.Errorf("Operational metric %s not found in registry", metricName)
			}
		})
	}

	// Test that removed metrics are NOT present
	removedMetrics := []string{
		"filesystem_size_bytes",
		"filesystem_free_bytes",
		"filesystem_available_bytes",
		"filesystem_used_bytes",
		"filesystem_inodes",
		"filesystem_inodes_free",
		"filesystem_inodes_used",
		"filesystem_readonly",
		"filesystem_device_error",
		"filesystem_exporter_last_collection_timestamp",
		"filesystem_exporter_collection_errors_total",
		"filesystem_collection_failed_total", // the old one
		"filesystem_collection_success_by_group_total",
		"filesystem_exporter_collection_duration_by_group_seconds",
	}

	for _, metricName := range removedMetrics {
		t.Run("removed_metric_"+metricName, func(t *testing.T) {
			metricFamilies, err := testutil.GatherAndCount(baseRegistry.GetRegistry(), metricName)
			if err != nil {
				t.Fatalf("Failed to gather metrics: %v", err)
			}

			if metricFamilies > 0 {
				t.Errorf("Removed metric %s still present in registry", metricName)
			}
		})
	}
}

func TestVolumeMetrics(t *testing.T) {
	// Create a base registry
	baseRegistry := promexporter_metrics.NewRegistry("test")

	// Create filesystem registry
	registry := NewFilesystemRegistry(baseRegistry)

	// Test volume metrics can be set
	registry.VolumeSizeGauge.With(prometheus.Labels{
		"device":      "sda1",
		"mount_point": "/",
		"volume":      "root",
	}).Set(1000000000)

	registry.VolumeAvailableGauge.With(prometheus.Labels{
		"device":      "sda1",
		"mount_point": "/",
		"volume":      "root",
	}).Set(500000000)

	registry.VolumeUsedRatioGauge.With(prometheus.Labels{
		"device":      "sda1",
		"mount_point": "/",
		"volume":      "root",
	}).Set(0.5)

	// Verify metrics were set
	metricFamilies, err := testutil.GatherAndCount(baseRegistry.GetRegistry())
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Should have at least the volume metrics we set
	if metricFamilies < 3 {
		t.Errorf("Expected at least 3 metric families, got %d", metricFamilies)
	}
}

func TestCollectionMetrics(t *testing.T) {
	// Create a base registry
	baseRegistry := promexporter_metrics.NewRegistry("test")

	// Create filesystem registry
	registry := NewFilesystemRegistry(baseRegistry)

	// Test collection metrics can be set
	registry.CollectionDuration.With(prometheus.Labels{
		"group":            "test",
		"interval_seconds": "60",
		"type":             "filesystem",
	}).Set(1.5)

	registry.CollectionSuccess.With(prometheus.Labels{
		"group":            "test",
		"interval_seconds": "60",
		"type":             "filesystem",
	}).Inc()

	registry.CollectionFailedCounter.With(prometheus.Labels{
		"group":            "test",
		"interval_seconds": "60",
		"type":             "filesystem",
	}).Inc()

	// Verify metrics were set
	metricFamilies, err := testutil.GatherAndCount(baseRegistry.GetRegistry())
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Should have at least the collection metrics we set
	if metricFamilies < 3 {
		t.Errorf("Expected at least 3 metric families, got %d", metricFamilies)
	}
}

func TestDirectoryMetrics(t *testing.T) {
	// Create a base registry
	baseRegistry := promexporter_metrics.NewRegistry("test")

	// Create filesystem registry
	registry := NewFilesystemRegistry(baseRegistry)

	// Test directory metrics can be set
	registry.DirectorySizeGauge.With(prometheus.Labels{
		"group":              "test",
		"directory":          "/tmp",
		"mode":               "du",
		"subdirectory_level": "0",
	}).Set(1024000)

	// Verify metrics were set
	metricFamilies, err := testutil.GatherAndCount(baseRegistry.GetRegistry())
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Should have at least the directory metric we set
	if metricFamilies < 1 {
		t.Errorf("Expected at least 1 metric family, got %d", metricFamilies)
	}
}
