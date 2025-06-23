package queue

import "time"

type Peer struct {
	intSet
	Timer *time.Timer // want to store pointer to timer so if messages hang or drop we can resend
	InFlight []int
}

func NewPeerQueue() *Peer {
	return &Peer{
		intSet: newIntSet(),
	}
}

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

type Messages struct {
	intSet
}

func NewMessagesQueue() *Messages {
	return &Messages{
		intSet: newIntSet(),
	}
}