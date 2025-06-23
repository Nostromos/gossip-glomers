package queue

import "time"

// Peer represents a peer node's message queue with retry capabilities.
// Embeds intSet for thread-safe integer set operations and adds
// timer-based retry logic for handling failed message transmissions.
type Peer struct {
	intSet
	
	// Timer manages retry timeouts for failed message transmissions
	// Stores pointer to timer so messages can be resent if they hang or drop
	Timer *time.Timer
	
	// InFlight tracks messages currently being transmitted to this peer
	// Used for acknowledgment handling and retry logic
	InFlight []int
}

// NewPeerQueue creates a new peer queue with initialized thread-safe integer set.
// Timer and InFlight fields are zero-initialized and managed by the gossip server.
func NewPeerQueue() *Peer {
	return &Peer{
		intSet: newIntSet(),
	}
}

// DrainBatch extracts up to 'limit' messages from the peer queue for transmission.
// Removes drained messages from the queue and returns them as a slice.
// Uses mutex locking to ensure thread-safe access during batch operations.
// Returns nil if queue is empty, otherwise returns slice of message IDs.
func (pq *Peer) DrainBatch(limit int) []int {
	pq.MU.Lock()
	defer pq.MU.Unlock()

	if len(pq.Values) == 0 {
		return nil
	}

	batchSize := limit
	if len(pq.Values) < limit {
		batchSize = len(pq.Values)
	}

	batch := make([]int, 0, batchSize)
	count := 0

	for k := range pq.Values {
		if count >= limit {
			break
		}

		batch = append(batch, k)
		delete(pq.Values, k)
		count++
	}

	return batch
}

// Messages represents the global message storage for the distributed system.
// Embeds intSet to provide thread-safe storage and retrieval of all seen messages
// across the entire gossip network. Used for deduplication and state management.
type Messages struct {
	intSet
}

// NewMessagesQueue creates a new global message queue with thread-safe integer set.
// This queue stores all messages seen by the node for deduplication and
// serves as the authoritative source of truth for broadcast state.
func NewMessagesQueue() *Messages {
	return &Messages{
		intSet: newIntSet(),
	}
}