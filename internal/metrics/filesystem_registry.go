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
	VolumeSizeGauge      *prometheus.GaugeVec
	VolumeAvailableGauge *prometheus.GaugeVec
	VolumeUsedRatioGauge *prometheus.GaugeVec

	// Directory metrics (documented)
	DirectorySizeGauge *prometheus.GaugeVec

	// Collection metrics (documented)
	CollectionDuration      *prometheus.GaugeVec
	CollectionSuccess       *prometheus.CounterVec
	CollectionFailedCounter *prometheus.CounterVec
	CollectionTotal         *prometheus.CounterVec

	// Additional operational metrics (used by collectors but not documented)
	CollectionIntervalGauge     *prometheus.GaugeVec
	CollectionTimestampGauge    *prometheus.GaugeVec
	DirectoriesFailedCounter    *prometheus.CounterVec
	DuLockWaitDurationGauge     *prometheus.GaugeVec
	DirectoriesProcessedCounter *prometheus.CounterVec

	// Operational metrics
	QueueDepthGauge          *prometheus.GaugeVec
	QueueWaitSecondsGauge    *prometheus.GaugeVec
	CollectionActiveGauge    *prometheus.GaugeVec
	CollectionSkippedCounter *prometheus.CounterVec
	GoroutineCountGauge      prometheus.Gauge

	// Per-job resource metrics (self-measurement)
	JobCPUUserSeconds       *prometheus.GaugeVec
	JobCPUSystemSeconds     *prometheus.GaugeVec
	JobMemoryAllocatedBytes *prometheus.GaugeVec
	JobMemoryPeakBytes      *prometheus.GaugeVec
	JobIOWaitSeconds        *prometheus.GaugeVec

	// Process-level resource metrics (self-measurement)
	ProcessCPUUserSecondsTotal   prometheus.Counter
	ProcessCPUSystemSecondsTotal prometheus.Counter
	ProcessMemoryAllocBytes      prometheus.Gauge
	ProcessMemorySysBytes        prometheus.Gauge
	ProcessNumGCTotal            prometheus.Gauge

	// Timeout metrics
	CollectionTimeoutSeconds *prometheus.GaugeVec
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
		CollectionTotal: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_collection_total",
				Help: "Total number of collections (successful and failed)",
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

		// Operational metrics
		QueueDepthGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_queue_depth",
				Help: "Current queue depth",
			},
			[]string{"queue_type"},
		),
		QueueWaitSecondsGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_queue_wait_seconds",
				Help: "Time jobs wait in queue",
			},
			[]string{"queue_type"},
		),
		CollectionActiveGauge: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_active",
				Help: "Currently running collection (1 if active, 0 if idle)",
			},
			[]string{"queue_type"},
		),
		CollectionSkippedCounter: promauto.With(baseRegistry.GetRegistry()).NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_collection_skipped_total",
				Help: "Total number of skipped collections",
			},
			[]string{"queue_type", "item_name", "reason"},
		),
		GoroutineCountGauge: promauto.With(baseRegistry.GetRegistry()).NewGauge(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_goroutines",
				Help: "Number of goroutines (for debugging)",
			},
		),

		// Per-job resource metrics
		JobCPUUserSeconds: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_job_cpu_user_seconds",
				Help: "User CPU time per job",
			},
			[]string{"job_type", "job_name"},
		),
		JobCPUSystemSeconds: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_job_cpu_system_seconds",
				Help: "System CPU time per job",
			},
			[]string{"job_type", "job_name"},
		),
		JobMemoryAllocatedBytes: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_job_memory_allocated_bytes",
				Help: "Memory allocated during job",
			},
			[]string{"job_type", "job_name"},
		),
		JobMemoryPeakBytes: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_job_memory_peak_bytes",
				Help: "Peak memory usage during job",
			},
			[]string{"job_type", "job_name"},
		),
		JobIOWaitSeconds: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_job_iowait_seconds",
				Help: "I/O wait time per job (if available)",
			},
			[]string{"job_type", "job_name"},
		),

		// Process-level resource metrics
		ProcessCPUUserSecondsTotal: promauto.With(baseRegistry.GetRegistry()).NewCounter(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_process_cpu_user_seconds_total",
				Help: "Total user CPU time",
			},
		),
		ProcessCPUSystemSecondsTotal: promauto.With(baseRegistry.GetRegistry()).NewCounter(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_process_cpu_system_seconds_total",
				Help: "Total system CPU time",
			},
		),
		ProcessMemoryAllocBytes: promauto.With(baseRegistry.GetRegistry()).NewGauge(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_process_memory_alloc_bytes",
				Help: "Current memory allocated",
			},
		),
		ProcessMemorySysBytes: promauto.With(baseRegistry.GetRegistry()).NewGauge(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_process_memory_sys_bytes",
				Help: "Memory obtained from OS",
			},
		),
		ProcessNumGCTotal: promauto.With(baseRegistry.GetRegistry()).NewGauge(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_process_num_gc",
				Help: "Number of GC cycles",
			},
		),

		// Timeout metrics
		CollectionTimeoutSeconds: promauto.With(baseRegistry.GetRegistry()).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_timeout_seconds",
				Help: "Configured timeout per item",
			},
			[]string{"item_name", "item_type"},
		),
	}

	// Add metric metadata for UI (only documented metrics)
	filesystem.AddMetricInfo("filesystem_exporter_volume_size_bytes", "Total size of volume in bytes", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_available_bytes", "Available space on volume in bytes", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_volume_used_ratio", "Ratio of used space on volume (0.0 to 1.0)", []string{"volume", "mount_point", "device"})
	filesystem.AddMetricInfo("filesystem_exporter_directory_size_bytes", "Size of directory in bytes", []string{"group", "directory", "mode", "subdirectory_level"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_duration_seconds", "Duration of collection in seconds", []string{"group", "interval_seconds", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_success_total", "Total number of successful collections", []string{"group", "interval_seconds", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_failed_total", "Total number of failed collections", []string{"group", "interval_seconds", "type"})
	filesystem.AddMetricInfo("filesystem_exporter_collection_total", "Total number of collections (successful and failed)", []string{"group", "interval_seconds", "type"})

	return filesystem
}
