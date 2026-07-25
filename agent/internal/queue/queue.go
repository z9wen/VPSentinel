package queue

import "github.com/z9wen/VPSentinel/protocol"

// SampleQueue is a bounded in-memory buffer. Dropping the oldest sample at
// capacity prevents an extended control-plane outage from exhausting memory.
type SampleQueue struct {
	capacity int
	samples  []protocol.MetricSample
}

func New(capacity int) *SampleQueue {
	if capacity <= 0 {
		panic("queue capacity must be greater than zero")
	}
	return &SampleQueue{capacity: capacity, samples: make([]protocol.MetricSample, 0, capacity)}
}

// Push reports whether the oldest sample had to be dropped.
func (q *SampleQueue) Push(sample protocol.MetricSample) bool {
	dropped := len(q.samples) == q.capacity
	if dropped {
		copy(q.samples, q.samples[1:])
		q.samples[len(q.samples)-1] = sample
		return true
	}
	q.samples = append(q.samples, sample)
	return false
}

func (q *SampleQueue) Snapshot(limit int) []protocol.MetricSample {
	if limit <= 0 || limit > len(q.samples) {
		limit = len(q.samples)
	}
	result := make([]protocol.MetricSample, limit)
	copy(result, q.samples[:limit])
	return result
}

// Remove deletes only acknowledged sample IDs. This is safe when samples were
// added or old samples were evicted while a batch was in flight.
func (q *SampleQueue) Remove(sampleIDs []string) int {
	if len(sampleIDs) == 0 {
		return 0
	}

	acknowledged := make(map[string]struct{}, len(sampleIDs))
	for _, sampleID := range sampleIDs {
		acknowledged[sampleID] = struct{}{}
	}

	original := q.samples
	kept := original[:0]
	for _, sample := range original {
		if _, found := acknowledged[sample.SampleID]; found {
			continue
		}
		kept = append(kept, sample)
	}
	removed := len(original) - len(kept)
	clear(original[len(kept):])
	q.samples = kept
	return removed
}

func (q *SampleQueue) Len() int {
	return len(q.samples)
}
