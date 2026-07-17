package ingest

import (
	"context"
	"sync"
	"time"

	"github.com/z9wen/VPSentinel/protocol"
)

type Store interface {
	StoreBatch(context.Context, string, protocol.MetricBatch, time.Time) (int, error)
	Ready(context.Context) error
}

type AcceptedBatch struct {
	NodeID     string
	Batch      protocol.MetricBatch
	ReceivedAt time.Time
}

// MemoryStore supports local development until the PostgreSQL repository is
// connected. It is bounded so a forgotten development server cannot grow
// without limit.
type MemoryStore struct {
	mu       sync.RWMutex
	capacity int
	batches  []AcceptedBatch
	seen     map[string]struct{}
}

func NewMemoryStore(capacity int) *MemoryStore {
	if capacity <= 0 {
		panic("memory store capacity must be greater than zero")
	}
	return &MemoryStore{
		capacity: capacity,
		batches:  make([]AcceptedBatch, 0, capacity),
		seen:     make(map[string]struct{}, capacity),
	}
}

func (s *MemoryStore) StoreBatch(
	_ context.Context,
	nodeID string,
	batch protocol.MetricBatch,
	receivedAt time.Time,
) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acceptedSamples := make([]protocol.MetricSample, 0, len(batch.Samples))
	for _, sample := range batch.Samples {
		key := sampleKey(nodeID, sample.SampleID)
		if _, duplicate := s.seen[key]; duplicate {
			continue
		}
		s.seen[key] = struct{}{}
		acceptedSamples = append(acceptedSamples, sample)
	}
	if len(acceptedSamples) == 0 {
		return 0, nil
	}
	batch.Samples = acceptedSamples

	accepted := AcceptedBatch{NodeID: nodeID, Batch: batch, ReceivedAt: receivedAt}
	if len(s.batches) == s.capacity {
		for _, sample := range s.batches[0].Batch.Samples {
			delete(s.seen, sampleKey(s.batches[0].NodeID, sample.SampleID))
		}
		copy(s.batches, s.batches[1:])
		s.batches[len(s.batches)-1] = accepted
		return len(acceptedSamples), nil
	}
	s.batches = append(s.batches, accepted)
	return len(acceptedSamples), nil
}

func (s *MemoryStore) Ready(context.Context) error {
	return nil
}

func (s *MemoryStore) Batches() []AcceptedBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]AcceptedBatch, len(s.batches))
	copy(result, s.batches)
	return result
}

func sampleKey(nodeID, sampleID string) string {
	return nodeID + "\x00" + sampleID
}
