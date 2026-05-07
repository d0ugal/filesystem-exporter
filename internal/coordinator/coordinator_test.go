package coordinator

import (
	"context"
	"runtime"
	"testing"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"

	promexporter_config "github.com/d0ugal/promexporter/config"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
)

// TestCoordinator_StopsOnContextCancel locks in the graceful-shutdown fix:
// Coordinator.Start spawns goroutines (workers, scheduler tickers,
// goroutine-count updater, queue-depth updater) under the supplied context.
// Cancelling that context must cause every goroutine to exit, so app.Run()
// can manage the coordinator's lifecycle via app.WithCollector.
func TestCoordinator_StopsOnContextCancel(t *testing.T) {
	cfg := &config.Config{
		BaseConfig: promexporter_config.BaseConfig{
			Metrics: promexporter_config.MetricsConfig{
				Collection: promexporter_config.CollectionConfig{
					DefaultInterval:    promexporter_config.Duration{Duration: 30 * time.Second},
					DefaultIntervalSet: true,
				},
			},
		},
	}

	metricsRegistry := promexporter_metrics.NewRegistry("filesystem_exporter_test_info")
	filesystemMetrics := metrics.NewFilesystemRegistry(metricsRegistry)
	coord := NewCoordinator(cfg, filesystemMetrics, nil)

	initial := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	coord.Start(ctx)

	// Allow goroutines to spawn.
	time.Sleep(200 * time.Millisecond)

	if runtime.NumGoroutine() <= initial {
		t.Fatalf("expected coordinator to spawn goroutines, initial=%d, after Start=%d",
			initial, runtime.NumGoroutine())
	}

	cancel()
	coord.Stop()

	// Wait up to 5s for the goroutine count to fall back close to baseline.
	// We allow a small slack (initial+2) because runtime metrics, GC workers,
	// and similar shared infrastructure add unpredictable jitter.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= initial+2 {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("coordinator goroutines did not exit after context cancellation: initial=%d, current=%d",
		initial, runtime.NumGoroutine())
}
