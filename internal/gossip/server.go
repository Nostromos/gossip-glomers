package gossip

import (
	// --- Standard Lib
	"sync"
	"sync/atomic"
	"time"

	// --- Internal Lib ---
	"maelstrom-broadcast/internal/queue"

	// --- Third Party ---
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Server struct {
	Node     *maelstrom.Node
	Messages *queue.Messages
	Pending  map[string]*queue.Peer
	Counter  atomic.Uint64
	initOnce sync.Once
	GossipInterval time.Duration
	RetryTimeout time.Duration
}

func NewServer(n *maelstrom.Node) *Server {
	return &Server{
		Node: n,
		Messages: &queue.Messages{
			Values: make(map[int]struct{})},
		GossipInterval: 100 * time.Millisecond, 
		RetryTimeout: 100 * time.Millisecond,
		// Pending: make(map[string]*queue.Peer), // Not clear to me that initializing an empty queue of peers prior to receiving them is smart or worthwhile.
	}
}

func (s *Server) HandlePeerQueues(node *maelstrom.Node, pending map[string]*queue.Peer) {
	ticker := time.NewTicker(s.GossipInterval)

	for range ticker.C {
		for peerID, pq := range pending {
			queue.drainAndSend(node, peerID, pq)
		}
	}
}


func (s *Server) DrainAndSend(node *maelstrom.Node, peer string, pq *queue.Peer) {
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
	pq.Timer = time.AfterFunc(s.RetryTimeout, func() {
		for _, v := range batch {
			pq.Values[v] = struct{}{}
		}

		pq.Timer = nil
	})

	pq.MU.Unlock()
}
