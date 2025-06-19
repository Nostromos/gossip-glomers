package queue

import (
	// --- Standard Lib ---
	"sync"
	"time"

	// --- Third Party ---
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Peer struct {
	MU     sync.RWMutex
	Values map[int]struct{}
	Timer  *time.Timer // want to store pointer to timer so if messages hang or drop we can resend
}

func (pq *Peer) Push(v int) bool {
	pq.MU.Lock()
	defer pq.MU.Unlock()

	if _, exists := pq.Values[v]; exists {
		return false
	}

	pq.Values[v] = struct{}{}
	return true
}

func (pq *Peer) PopOne() (int, bool) {
	pq.MU.Lock()
	defer pq.MU.Unlock()

	for v := range pq.Values {
		delete(pq.Values, v)
		return v, true
	}
	return 0, false
}

func (pq *Peer) Drain() []int {
	pq.MU.Lock()
	defer pq.MU.Unlock()

	copy := make([]int, 0, len(pq.Values))
	for k := range pq.Values {
		copy = append(copy, k)
		delete(pq.Values, k)
	}

	return copy
}

func HandlePeerQueues(node *maelstrom.Node, pending map[string]*Peer) {
	ticker := time.NewTicker(100 * time.Millisecond)

	for range ticker.C {
		for peerID, pq := range pending {
			drainAndSend(node, peerID, pq)
		}
	}
}

func drainAndSend(node *maelstrom.Node, peer string, pq *Peer) {
	batch := pq.Drain()

	if len(batch) == 0 {
		return
	}
	body := map[string]any{
		"type":     "delta",
		"messages": batch,
	}
	node.Send(peer, body)
	pq.MU.Lock()
	if pq.Timer != nil {
		pq.Timer.Stop()
	}
	pq.Timer = time.AfterFunc(200*time.Millisecond, func() {
		pq.MU.Lock()
		for _, v := range batch {
			pq.Values[v] = struct{}{}
		}

		pq.Timer = nil
		pq.MU.Unlock()
	})

	pq.MU.Unlock()
}
