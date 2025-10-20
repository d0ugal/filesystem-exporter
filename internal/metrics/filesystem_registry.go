package metrics

import (
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// FilesystemRegistry wraps the promexporter registry with filesystem-specific metrics
type FilesystemRegistry struct {
	*promexporter_metrics.Registry

	// Filesystem metrics
	FilesystemSizeBytes      *prometheus.GaugeVec
	FilesystemFreeBytes      *prometheus.GaugeVec
	FilesystemAvailableBytes *prometheus.GaugeVec
	FilesystemUsedBytes      *prometheus.GaugeVec
	FilesystemInodesTotal    *prometheus.GaugeVec
	FilesystemInodesFree     *prometheus.GaugeVec
	FilesystemInodesUsed     *prometheus.GaugeVec
	FilesystemReadOnly       *prometheus.GaugeVec
	FilesystemDeviceError    *prometheus.GaugeVec

	// Directory metrics (only DirectorySizeGauge is actually used)

	// Collection metrics
	LastCollectionTime          *prometheus.GaugeVec
	CollectionDuration          *prometheus.GaugeVec
	CollectionErrors            *prometheus.CounterVec
	CollectionSuccess           *prometheus.CounterVec
	CollectionFailedCounter     *prometheus.CounterVec
	CollectionSuccessCounter    *prometheus.CounterVec
	CollectionIntervalGauge     *prometheus.GaugeVec
	CollectionDurationGauge     *prometheus.GaugeVec
	CollectionTimestampGauge    *prometheus.GaugeVec
	DirectoriesFailedCounter    *prometheus.CounterVec
	DuLockWaitDurationGauge     *prometheus.GaugeVec
	DirectorySizeGauge          *prometheus.GaugeVec
	DirectoriesProcessedCounter *prometheus.CounterVec
	VolumeSizeGauge             *prometheus.GaugeVec
	VolumeAvailableGauge        *prometheus.GaugeVec
	VolumeUsedRatioGauge        *prometheus.GaugeVec
}

// NewFilesystemRegistry creates a new filesystem metrics registry
func NewFilesystemRegistry(baseRegistry *promexporter_metrics.Registry) *FilesystemRegistry {
	filesystem := &FilesystemRegistry{
		Registry: baseRegistry,

		// Filesystem metrics
		FilesystemSizeBytes: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_size_bytes",
				Help: "Filesystem size in bytes",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemFreeBytes: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_free_bytes",
				Help: "Filesystem free space in bytes",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemAvailableBytes: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_available_bytes",
				Help: "Filesystem available space in bytes",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemUsedBytes: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_used_bytes",
				Help: "Filesystem used space in bytes",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemInodesTotal: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_inodes",
				Help: "Filesystem total inodes",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemInodesFree: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_inodes_free",
				Help: "Filesystem free inodes",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemInodesUsed: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_inodes_used",
				Help: "Filesystem used inodes",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemReadOnly: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_readonly",
				Help: "Filesystem is read-only (1) or writable (0)",
			},
			[]string{"device", "mountpoint", "fstype"},
		),
		FilesystemDeviceError: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_device_error",
				Help: "Filesystem device error (1) or no error (0)",
			},
			[]string{"device", "mountpoint", "fstype"},
		),

		// Directory metrics (only DirectorySizeGauge is actually used)

		// Collection metrics
		LastCollectionTime: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_last_collection_timestamp",
				Help: "Timestamp of last successful collection",
			},
			[]string{"group", "interval_seconds", "type"},
		),
		CollectionDuration: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_duration_seconds",
				Help: "Duration of last collection in seconds",
			},
			[]string{"group", "interval_seconds", "type"},
		),
		CollectionErrors: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_collection_errors_total",
				Help: "Total number of collection errors",
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
				Name: "filesystem_collection_failed_total",
				Help: "Total number of failed collections",
			},
			[]string{"collector", "group", "interval"},
		),
		CollectionSuccessCounter: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_collection_success_by_group_total",
				Help: "Total number of successful collections by group",
			},
			[]string{"collector", "group", "interval"},
		),
		CollectionIntervalGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_interval_seconds",
				Help: "Collection interval in seconds",
			},
			[]string{"group", "type"},
		),
		CollectionDurationGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_duration_by_group_seconds",
				Help: "Duration of collection by group in seconds",
			},
			[]string{"group", "interval_seconds", "type"},
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
		DirectorySizeGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_directory_size_bytes",
				Help: "Directory size in bytes",
			},
			[]string{"group", "directory", "mode", "subdirectory_level"},
		),
		DirectoriesProcessedCounter: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_directories_processed_total",
				Help: "Total number of directories processed",
			},
			[]string{"group", "method"},
		),
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
	}

	// Add metric metadata for UI
	filesystem.AddMetricInfo("filesystem_exporter_info", "Information about the filesystem exporter", []string{"version", "commit", "build_date"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_size_bytes", "Total size of volume in bytes", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_available_bytes", "Available space on volume in bytes", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_used_ratio", "Ratio of used space on volume (0.0 to 1.0)", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_directory_size_bytes", "Size of directory in bytes", []string{"group", "directory", "mode", "subdirectory_level"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_duration_seconds", "Duration of collection in seconds", []string{"group", "interval_seconds", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_timestamp", "Timestamp of last collection", []string{"group", "interval_seconds", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_interval_seconds", "Configured collection interval in seconds", []string{"group", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_success_total", "Total number of successful collections", []string{"type", "group", "interval_seconds"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_failed_total", "Total number of failed collections", []string{"type", "group", "interval_seconds"})
	filesystem.AddMetricInfo("filesystem_exporter_directories_processed_total", "Total number of directories processed", []string{"group", "method"})
	filesystem.AddMetricInfo("filesystem_exporter_directories_failed_total", "Total number of directories that failed to process", []string{"group", "reason"})
	filesystem.AddMetricInfo("filesystem_exporter_du_lock_wait_duration_seconds", "Time spent waiting for du mutex lock in seconds", []string{"group", "path"})

	return filesystem
}
