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
}

func NewServer(n *maelstrom.Node) *Server {
	return &Server{
		Node: n,
		Messages: queue.NewMessagesQueue(),
		GossipInterval: 100 * time.Millisecond,
		RetryTimeout:   100 * time.Millisecond,
	}
}

func (s *Server) HandlePeerQueues() error {
	ticker := time.NewTicker(s.GossipInterval)

	for range ticker.C {
		for peerID, pq := range s.Pending {
			batch := pq.GetSlice()
			if len(batch) == 0 { 
				continue 
			} else {
				resp := protocol.DeltaReq{
					Type: "delta",
					Messages: batch,
				}
				s.Node.Send(peerID, resp)
			}
		}
	}
	return nil
}
