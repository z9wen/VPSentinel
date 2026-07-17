package collector

import (
	"context"
	"slices"
	"testing"
)

func TestCollectReturnsPartialSampleWhenDiskCollectorFails(t *testing.T) {
	collector := NewHostCollector("/path-that-does-not-exist-vpsentinel-test")
	sample, err := collector.Collect(context.Background())
	if err == nil {
		t.Fatal("Collect returned nil error, want disk collector error")
	}
	if sample.RecordedAt.IsZero() {
		t.Fatal("RecordedAt is zero")
	}
	if !slices.Contains(sample.FailedCollectors, "disk") {
		t.Fatalf("FailedCollectors = %v, want disk", sample.FailedCollectors)
	}
	if sample.MemoryTotalBytes == nil {
		t.Fatal("memory metric is missing; want successful collectors preserved")
	}
}
