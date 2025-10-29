package metrics

import (
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// FilesystemRegistry wraps the promexporter registry with filesystem-specific metrics
type FilesystemRegistry struct {
	*promexporter_metrics.Registry

	// Volume metrics (documented)
	VolumeSizeGauge             *prometheus.GaugeVec
	VolumeAvailableGauge        *prometheus.GaugeVec
	VolumeUsedRatioGauge        *prometheus.GaugeVec

	// Directory metrics (documented)
	DirectorySizeGauge          *prometheus.GaugeVec

	// Collection metrics (documented)
	CollectionDuration          *prometheus.GaugeVec
	CollectionSuccess           *prometheus.CounterVec
	CollectionFailedCounter     *prometheus.CounterVec

	// Additional operational metrics (used by collectors but not documented)
	CollectionIntervalGauge     *prometheus.GaugeVec
	CollectionTimestampGauge    *prometheus.GaugeVec
	DirectoriesFailedCounter    *prometheus.CounterVec
	DuLockWaitDurationGauge     *prometheus.GaugeVec
	DirectoriesProcessedCounter *prometheus.CounterVec
}

// NewFilesystemRegistry creates a new filesystem metrics registry
func NewFilesystemRegistry(baseRegistry *promexporter_metrics.Registry) *FilesystemRegistry {
	filesystem := &FilesystemRegistry{
		Registry: baseRegistry,

		// Volume metrics (documented)
		VolumeSizeGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_volume_size_bytes",
				Help: "Volume size in bytes",
			},
			[]string{"device", "mount_point", "volume"},
		),
		VolumeAvailableGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_volume_available_bytes",
				Help: "Volume available space in bytes",
			},
			[]string{"device", "mount_point", "volume"},
		),
		VolumeUsedRatioGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_volume_used_ratio",
				Help: "Volume used space ratio (0-1)",
			},
			[]string{"device", "mount_point", "volume"},
		),

		// Directory metrics (documented)
		DirectorySizeGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_directory_size_bytes",
				Help: "Directory size in bytes",
			},
			[]string{"group", "directory", "mode", "subdirectory_level"},
		),

		// Collection metrics (documented)
		CollectionDuration: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_duration_seconds",
				Help: "Duration of last collection in seconds",
			},
			[]string{"group", "interval_seconds", "type"},
		),
		CollectionSuccess: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_collection_success_total",
				Help: "Total number of successful collections",
			},
			[]string{"group", "interval_seconds", "type"},
		),
		CollectionFailedCounter: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_collection_failed_total",
				Help: "Total number of failed collections",
			},
			[]string{"group", "interval_seconds", "type"},
		),

		// Additional operational metrics (used by collectors but not documented)
		CollectionIntervalGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_interval_seconds",
				Help: "Collection interval in seconds",
			},
			[]string{"group", "type"},
		),
		CollectionTimestampGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_timestamp",
				Help: "Timestamp of collection",
			},
			[]string{"group", "interval_seconds", "type"},
		),
		DirectoriesFailedCounter: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_directories_failed_total",
				Help: "Total number of failed directory operations",
			},
			[]string{"group", "reason"},
		),
		DuLockWaitDurationGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_du_lock_wait_duration_seconds",
				Help: "Duration waiting for du lock",
			},
			[]string{"group", "path"},
		),
		DirectoriesProcessedCounter: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_directories_processed_total",
				Help: "Total number of directories processed",
			},
			[]string{"group", "method"},
		),
	}

	// Add metric metadata for UI (only documented metrics)
	filesystem.AddMetricInfo("filesystem_exporter_info", "Information about the filesystem exporter", []string{"version", "commit", "build_date"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_size_bytes", "Total size of volume in bytes", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_available_bytes", "Available space on volume in bytes", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_used_ratio", "Ratio of used space on volume (0.0 to 1.0)", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_directory_size_bytes", "Size of directory in bytes", []string{"group", "directory", "mode", "subdirectory_level"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_duration_seconds", "Duration of collection in seconds", []string{"group", "interval_seconds", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_success_total", "Total number of successful collections", []string{"group", "interval_seconds", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_failed_total", "Total number of failed collections", []string{"group", "interval_seconds", "type"})

	return filesystem
}
