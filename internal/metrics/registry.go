package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricInfo contains information about a metric for the UI
type MetricInfo struct {
	Name         string
	Help         string
	Labels       []string
	ExampleValue string
}

type Registry struct {
	// Version info metric
	versionInfoGauge *prometheus.GaugeVec

	// Volume metrics
	volumeSizeGauge      *prometheus.GaugeVec
	volumeAvailableGauge *prometheus.GaugeVec
	volumeUsedRatioGauge *prometheus.GaugeVec

	// Directory metrics
	directorySizeGauge *prometheus.GaugeVec

	// Collection metrics
	collectionDurationGauge  *prometheus.GaugeVec
	collectionTimestampGauge *prometheus.GaugeVec
	collectionIntervalGauge  *prometheus.GaugeVec
	collectionSuccessCounter *prometheus.CounterVec
	collectionFailedCounter  *prometheus.CounterVec

	// Processing metrics
	directoriesProcessedCounter *prometheus.CounterVec
	directoriesFailedCounter    *prometheus.CounterVec

	// Lock waiting metrics
	duLockWaitDurationGauge *prometheus.GaugeVec

	// The underlying Prometheus registry
	registry *prometheus.Registry

	// Metric information for UI
	metricInfo []MetricInfo
}

// addMetricInfo adds metric information to the registry
func (r *Registry) addMetricInfo(name, help string, labels []string) {
	r.metricInfo = append(r.metricInfo, MetricInfo{
		Name:         name,
		Help:         help,
		Labels:       labels,
		ExampleValue: "",
	})
}

func NewRegistry() *Registry {
	registry := prometheus.NewRegistry()
	factory := promauto.With(registry)

	r := &Registry{
		versionInfoGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_info",
				Help: "Information about the filesystem exporter",
			},
			[]string{"version", "commit", "build_date"},
		),
		volumeSizeGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_volume_size_bytes",
				Help: "Total size of volume in bytes",
			},
			[]string{"volume", "mount_point", "device"},
		),
		volumeAvailableGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_volume_available_bytes",
				Help: "Available space on volume in bytes",
			},
			[]string{"volume", "mount_point", "device"},
		),
		volumeUsedRatioGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_volume_used_ratio",
				Help: "Ratio of used space on volume (0.0 to 1.0)",
			},
			[]string{"volume", "mount_point", "device"},
		),
		directorySizeGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_directory_size_bytes",
				Help: "Size of directory in bytes",
			},
			[]string{"group", "directory", "mode", "subdirectory_level"},
		),
		collectionDurationGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_duration_seconds",
				Help: "Duration of collection in seconds",
			},
			[]string{"type", "group", "interval_seconds"},
		),
		collectionTimestampGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_timestamp",
				Help: "Timestamp of last collection",
			},
			[]string{"type", "group", "interval_seconds"},
		),
		collectionIntervalGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_collection_interval_seconds",
				Help: "Configured collection interval in seconds",
			},
			[]string{"type", "group"},
		),
		collectionSuccessCounter: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_collection_success_total",
				Help: "Total number of successful collections",
			},
			[]string{"type", "group", "interval_seconds"},
		),
		collectionFailedCounter: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_collection_failed_total",
				Help: "Total number of failed collections",
			},
			[]string{"type", "group", "interval_seconds"},
		),
		directoriesProcessedCounter: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_directories_processed_total",
				Help: "Total number of directories processed",
			},
			[]string{"group", "mode"},
		),
		directoriesFailedCounter: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "filesystem_exporter_directories_failed_total",
				Help: "Total number of directories that failed to process",
			},
			[]string{"group", "mode"},
		),
		duLockWaitDurationGauge: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "filesystem_exporter_du_lock_wait_duration_seconds",
				Help: "Time spent waiting for du mutex lock in seconds",
			},
			[]string{"group", "directory"},
		),
		registry: registry,
	}

	// Add metric information for UI
	r.addMetricInfo("filesystem_exporter_info", "Information about the filesystem exporter", []string{"version", "commit", "build_date"})
	r.addMetricInfo("filesystem_exporter_volume_size_bytes", "Total size of volume in bytes", []string{"volume", "mount_point", "device"})
	r.addMetricInfo("filesystem_exporter_volume_available_bytes", "Available space on volume in bytes", []string{"volume", "mount_point", "device"})
	r.addMetricInfo("filesystem_exporter_volume_used_ratio", "Ratio of used space on volume (0.0 to 1.0)", []string{"volume", "mount_point", "device"})
	r.addMetricInfo("filesystem_exporter_directory_size_bytes", "Size of directory in bytes", []string{"group", "directory", "mode", "subdirectory_level"})
	r.addMetricInfo("filesystem_exporter_collection_duration_seconds", "Duration of collection in seconds", []string{"type", "group", "interval_seconds"})
	r.addMetricInfo("filesystem_exporter_collection_timestamp", "Timestamp of last collection", []string{"type", "group", "interval_seconds"})
	r.addMetricInfo("filesystem_exporter_collection_interval_seconds", "Configured collection interval in seconds", []string{"type", "group"})
	r.addMetricInfo("filesystem_exporter_collection_success_total", "Total number of successful collections", []string{"type", "group", "interval_seconds"})
	r.addMetricInfo("filesystem_exporter_collection_failed_total", "Total number of failed collections", []string{"type", "group", "interval_seconds"})
	r.addMetricInfo("filesystem_exporter_directories_processed_total", "Total number of directories processed", []string{"group", "mode"})
	r.addMetricInfo("filesystem_exporter_directories_failed_total", "Total number of directories that failed to process", []string{"group", "mode"})
	r.addMetricInfo("filesystem_exporter_du_lock_wait_duration_seconds", "Time spent waiting for du mutex lock in seconds", []string{"group", "directory"})

	return r
}

// GetRegistry returns the underlying Prometheus registry for HTTP exposure
func (r *Registry) GetRegistry() *prometheus.Registry {
	return r.registry
}

func (r *Registry) VolumeSizeGauge() *prometheus.GaugeVec {
	return r.volumeSizeGauge
}

func (r *Registry) VolumeAvailableGauge() *prometheus.GaugeVec {
	return r.volumeAvailableGauge
}

func (r *Registry) VolumeUsedRatioGauge() *prometheus.GaugeVec {
	return r.volumeUsedRatioGauge
}

func (r *Registry) DirectorySizeGauge() *prometheus.GaugeVec {
	return r.directorySizeGauge
}

func (r *Registry) CollectionDurationGauge() *prometheus.GaugeVec {
	return r.collectionDurationGauge
}

func (r *Registry) CollectionTimestampGauge() *prometheus.GaugeVec {
	return r.collectionTimestampGauge
}

func (r *Registry) CollectionIntervalGauge() *prometheus.GaugeVec {
	return r.collectionIntervalGauge
}

func (r *Registry) CollectionSuccessCounter() *prometheus.CounterVec {
	return r.collectionSuccessCounter
}

func (r *Registry) CollectionFailedCounter() *prometheus.CounterVec {
	return r.collectionFailedCounter
}

func (r *Registry) DirectoriesProcessedCounter() *prometheus.CounterVec {
	return r.directoriesProcessedCounter
}

func (r *Registry) DirectoriesFailedCounter() *prometheus.CounterVec {
	return r.directoriesFailedCounter
}

func (r *Registry) VersionInfoGauge() *prometheus.GaugeVec {
	return r.versionInfoGauge
}

func (r *Registry) DuLockWaitDurationGauge() *prometheus.GaugeVec {
	return r.duLockWaitDurationGauge
}

// GetMetricsInfo returns information about all metrics for the UI
func (r *Registry) GetMetricsInfo() []MetricInfo {
	return r.metricInfo
}
