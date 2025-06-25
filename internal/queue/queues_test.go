package queue

import (
	// --- Standard Lib ---
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerQueue(t *testing.T) {
	pq := NewPeerQueue()
	batch := pq.DrainBatch(10)

	if len(batch) > 0 {
		t.Errorf("Batch Size = %d, want 0", len(batch))
	}
}

func TestPeerQueue_DrainBatch(t *testing.T) {
	pq := NewPeerQueue()
	pq.Add(1)
	pq.Add(2)
	pq.Add(3)
	fmt.Println(pq)

	batch := pq.DrainBatch(2)
	assert.Len(t, batch, 2)
	assert.Contains(t, []int{1, 2, 3}, batch[0])
	assert.Contains(t, []int{1, 2, 3}, batch[1])

	// Verify items were removed
	assert.False(t, pq.Has(batch[0]))
	assert.False(t, pq.Has(batch[1]))
}

// Batch Size Limits:
func TestPeerQueue_DrainBatch_LimitRespected(t *testing.T) {
	pq := NewPeerQueue()
	for i := 0; i < 100; i++ {
			pq.Add(i)
	}

	batch := pq.DrainBatch(10)
	assert.Len(t, batch, 10)

	// Should have 90 remaining
	remaining := pq.GetSlice()
	assert.Len(t, remaining, 90)
}

// Concurrency Tests:
func TestPeerQueue_ConcurrentAccess(t *testing.T) {
	pq := NewPeerQueue()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(val int) {
					defer wg.Done()
					pq.Add(val)
			}(i)
	}

	// Concurrent drains
	batches := make([][]int, 5)
	for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
					defer wg.Done()
					batches[idx] = pq.DrainBatch(3)
			}(i)
	}

	wg.Wait()

	// Verify no data races and reasonable results
	totalDrained := 0
	for _, batch := range batches {
			totalDrained += len(batch)
	}
	assert.LessOrEqual(t, totalDrained, 10)
}