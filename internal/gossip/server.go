package gossip

import (
	// --- Standard Lib
	"sync"
	"sync/atomic"
	"time"

	// --- Internal Lib ---
	"maelstrom-broadcast/internal/protocol"
	"maelstrom-broadcast/internal/queue"

	// --- Third Party ---
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Server struct {
	Node           *maelstrom.Node
	Messages       *queue.Messages
	Pending        map[string]*queue.Peer
	Counter        atomic.Uint64
	initOnce       sync.Once
	GossipInterval time.Duration
	RetryTimeout   time.Duration
	GossipMax      int
}

func NewServer(n *maelstrom.Node) *Server {
	return &Server{
		Node:           n,
		Messages:       queue.NewMessagesQueue(),
		GossipInterval: 50 * time.Millisecond,
		RetryTimeout:   100 * time.Millisecond,
		GossipMax:      128,
	}
}

func (s *Server) HandlePeerQueues() error {
	ticker := time.NewTicker(s.GossipInterval)

	for range ticker.C {
		for peerID, pq := range s.Pending {
			
			if pq.InFlight == nil {
				batch := pq.DrainBatch(s.GossipMax)
				pq.MU.Lock()
				pq.InFlight = batch
				pq.MU.Unlock()
			} 

			if len(pq.InFlight) == 0 {
				continue
			}

			pq.MU.Lock()
			s.Node.Send(peerID, protocol.DeltaReq{
				Type:     "delta",
				Messages: pq.InFlight,
			})
			pq.MU.Unlock()
		}
	}
	return nil
}
