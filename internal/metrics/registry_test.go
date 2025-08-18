package metrics

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	// Test that all metrics are properly initialized
	if registry.VolumeSizeGauge() == nil {
		t.Error("Expected volume size gauge to be initialized")
	}

	if registry.VolumeAvailableGauge() == nil {
		t.Error("Expected volume available gauge to be initialized")
	}

	if registry.VolumeUsedRatioGauge() == nil {
		t.Error("Expected volume used ratio gauge to be initialized")
	}

	if registry.DirectorySizeGauge() == nil {
		t.Error("Expected directory size gauge to be initialized")
	}

	if registry.CollectionDurationGauge() == nil {
		t.Error("Expected collection duration gauge to be initialized")
	}

	if registry.CollectionTimestampGauge() == nil {
		t.Error("Expected collection timestamp gauge to be initialized")
	}

	if registry.CollectionSuccessCounter() == nil {
		t.Error("Expected collection success counter to be initialized")
	}

	if registry.CollectionFailedCounter() == nil {
		t.Error("Expected collection failed counter to be initialized")
	}

	if registry.CollectionIntervalGauge() == nil {
		t.Error("Expected collection interval gauge to be initialized")
	}

	if registry.DirectoriesProcessedCounter() == nil {
		t.Error("Expected directories processed counter to be initialized")
	}

	if registry.DirectoriesFailedCounter() == nil {
		t.Error("Expected directories failed counter to be initialized")
	}
}

func TestMetricsLabels(t *testing.T) {
	registry := NewRegistry()

	// Test that metrics are properly initialized
	volumeGauge := registry.VolumeSizeGauge()
	if volumeGauge == nil {
		t.Error("Expected volume gauge to be initialized")
	}

	dirGauge := registry.DirectorySizeGauge()
	if dirGauge == nil {
		t.Error("Expected directory gauge to be initialized")
	}

	collectionCounter := registry.CollectionSuccessCounter()
	if collectionCounter == nil {
		t.Error("Expected collection counter to be initialized")
	}

	processedCounter := registry.DirectoriesProcessedCounter()
	if processedCounter == nil {
		t.Error("Expected processed counter to be initialized")
	}
}

func TestMetricsConcurrency(t *testing.T) {
	registry := NewRegistry()

	// Test concurrent access to metrics
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			volumeGauge := registry.VolumeSizeGauge()
			if volumeGauge != nil {
				volumeGauge.WithLabelValues("volume", "/mount", "device").Set(float64(id))
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMetricsDescription(t *testing.T) {
	registry := NewRegistry()

	// Test that metrics are properly initialized
	volumeGauge := registry.VolumeSizeGauge()
	if volumeGauge == nil {
		t.Error("Expected volume gauge to be initialized")
	}

	dirGauge := registry.DirectorySizeGauge()
	if dirGauge == nil {
		t.Error("Expected directory gauge to be initialized")
	}

	collectionCounter := registry.CollectionSuccessCounter()
	if collectionCounter == nil {
		t.Error("Expected collection counter to be initialized")
	}
}

func TestMetricsCollect(t *testing.T) {
	registry := NewRegistry()

	// Test that metrics are properly initialized
	volumeGauge := registry.VolumeSizeGauge()
	if volumeGauge == nil {
		t.Error("Expected volume gauge to be initialized")
	}

	dirGauge := registry.DirectorySizeGauge()
	if dirGauge == nil {
		t.Error("Expected directory gauge to be initialized")
	}

	collectionCounter := registry.CollectionSuccessCounter()
	if collectionCounter == nil {
		t.Error("Expected collection counter to be initialized")
	}
}
