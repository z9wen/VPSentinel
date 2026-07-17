package queue

import (
	"testing"

	"github.com/z9wen/VPSentinel/protocol"
)

func TestQueueDropsOldestSampleAtCapacity(t *testing.T) {
	q := New(2)
	q.Push(protocol.MetricSample{SampleID: "one"})
	q.Push(protocol.MetricSample{SampleID: "two"})
	if dropped := q.Push(protocol.MetricSample{SampleID: "three"}); !dropped {
		t.Fatal("Push reported no dropped sample, want one")
	}

	samples := q.Snapshot(0)
	if len(samples) != 2 || samples[0].SampleID != "two" || samples[1].SampleID != "three" {
		t.Fatalf("Snapshot = %#v, want sample IDs [two, three]", samples)
	}
}

func TestSnapshotDoesNotExposeQueueStorage(t *testing.T) {
	q := New(1)
	q.Push(protocol.MetricSample{SampleID: "original"})
	snapshot := q.Snapshot(0)
	snapshot[0].SampleID = "changed"
	if q.Snapshot(0)[0].SampleID != "original" {
		t.Fatal("modifying snapshot changed queue storage")
	}
}

func TestRemoveOnlyAcknowledgedSamples(t *testing.T) {
	q := New(3)
	q.Push(protocol.MetricSample{SampleID: "one"})
	q.Push(protocol.MetricSample{SampleID: "two"})
	q.Push(protocol.MetricSample{SampleID: "three"})

	if removed := q.Remove([]string{"one", "two"}); removed != 2 {
		t.Fatalf("Remove removed %d samples, want 2", removed)
	}
	remaining := q.Snapshot(0)
	if len(remaining) != 1 || remaining[0].SampleID != "three" {
		t.Fatalf("remaining samples = %#v, want [three]", remaining)
	}
}

func TestRemoveDoesNotDeleteReplacementForEvictedInFlightSample(t *testing.T) {
	q := New(2)
	q.Push(protocol.MetricSample{SampleID: "in-flight"})
	q.Push(protocol.MetricSample{SampleID: "two"})
	q.Push(protocol.MetricSample{SampleID: "new"})

	q.Remove([]string{"in-flight"})
	remaining := q.Snapshot(0)
	if len(remaining) != 2 || remaining[0].SampleID != "two" || remaining[1].SampleID != "new" {
		t.Fatalf("remaining samples = %#v, want [two, new]", remaining)
	}
}
