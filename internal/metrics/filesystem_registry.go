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

	// Directory metrics
	DirectoryFileCount    *prometheus.GaugeVec
	DirectorySubdirCount  *prometheus.GaugeVec
	DirectoryLastModified *prometheus.GaugeVec
	DirectoryAccessible   *prometheus.GaugeVec
	DirectoryError        *prometheus.GaugeVec

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

		// Directory metrics
		DirectoryFileCount: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "directory_files",
				Help: "Number of files in directory",
			},
			[]string{"path"},
		),
		DirectorySubdirCount: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "directory_subdirs",
				Help: "Number of subdirectories in directory",
			},
			[]string{"path"},
		),
		DirectoryLastModified: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "directory_last_modified_timestamp",
				Help: "Directory last modified timestamp",
			},
			[]string{"path"},
		),
		DirectoryAccessible: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "directory_accessible",
				Help: "Directory is accessible (1) or not (0)",
			},
			[]string{"path"},
		),
		DirectoryError: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "directory_error",
				Help: "Directory error (1) or no error (0)",
			},
			[]string{"path"},
		),

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
			[]string{"group", "path", "method", "level"},
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
	filesystem.AddMetricInfo("filesystem_size_bytes", "Total size of the filesystem in bytes", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_free_bytes", "Free space available on the filesystem in bytes", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_available_bytes", "Space available to non-root users on the filesystem in bytes", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_used_bytes", "Used space on the filesystem in bytes", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_inodes_total", "Total number of inodes on the filesystem", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_inodes_free", "Number of free inodes on the filesystem", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_inodes_used", "Number of used inodes on the filesystem", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_readonly", "Whether the filesystem is read-only (1) or writable (0)", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("filesystem_device_error", "Whether there is a device error on the filesystem (1) or not (0)", []string{"device", "mountpoint", "fstype"})
	filesystem.AddMetricInfo("directory_files", "Number of files in the directory", []string{"path"})
	filesystem.AddMetricInfo("directory_subdir_count", "Number of subdirectories in the directory", []string{"path"})
	filesystem.AddMetricInfo("directory_last_modified_timestamp", "Timestamp when the directory was last modified", []string{"path"})
	filesystem.AddMetricInfo("directory_accessible", "Whether the directory is accessible (1) or not (0)", []string{"path"})
	filesystem.AddMetricInfo("directory_error", "Whether there is an error accessing the directory (1) or not (0)", []string{"path"})
	filesystem.AddMetricInfo("filesystem_last_collection_timestamp", "Timestamp of the last successful collection", []string{"collector"})
	filesystem.AddMetricInfo("filesystem_collection_duration_seconds", "Duration of the last collection in seconds", []string{"collector"})
	filesystem.AddMetricInfo("filesystem_collection_errors_total", "Total number of collection errors", []string{"collector"})
	filesystem.AddMetricInfo("filesystem_collection_success_total", "Total number of successful collections", []string{"collector"})
	filesystem.AddMetricInfo("filesystem_collection_success_by_group_total", "Total number of successful collections by group", []string{"collector", "group", "interval"})
	filesystem.AddMetricInfo("filesystem_collection_duration_by_group_seconds", "Duration of collection by group in seconds", []string{"collector", "group", "interval"})

	return filesystem
}
