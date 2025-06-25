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

// Server wraps a Maelstrom node with distributed gossip functionality.
// It manages message storage, peer queues, and periodic gossip dissemination
// for implementing distributed broadcast and consensus protocols.
type Server struct {
	// Node is the underlying Maelstrom node for network communication
	Node *maelstrom.Node
	
	// Messages stores all seen messages across the distributed system
	Messages *queue.Messages
	
	// Pending maps peer node IDs to their respective message queues
	// for gossip dissemination and retry logic
	Pending map[string]*queue.Peer
	
	// Counter provides atomic unique ID generation combined with node ID
	Counter atomic.Uint64
	
	// initOnce ensures topology initialization happens only once
	initOnce sync.Once
	
	// GossipInterval controls how frequently gossip messages are sent
	GossipInterval time.Duration
	
	// RetryTimeout defines how long to wait before retrying failed messages
	RetryTimeout time.Duration
	
	// GossipMax limits the number of messages sent in each gossip batch
	GossipMax int
}

// NewServer creates a new gossip server wrapping the provided Maelstrom node.
// Initializes default timing parameters: 50ms gossip interval, 100ms retry timeout,
// and maximum batch size of 128 messages per gossip round.
func NewServer(n *maelstrom.Node) *Server {
	return &Server{
		Node:           n,
		Messages:       queue.NewMessagesQueue(),
		GossipInterval: 50 * time.Millisecond,
		RetryTimeout:   100 * time.Millisecond,
		GossipMax:      128,
	}
}

// HandlePeerQueues runs the background gossip dissemination loop.
// Periodically drains peer queues and sends delta messages to propagate
// new messages throughout the distributed system. Manages in-flight messages
// and retry logic for failed transmissions.
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
