package gossip

import (
	// --- Standard Lib
	"sync"
	"sync/atomic"

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
}

func NewServer(n *maelstrom.Node) *Server {
	return &Server{
		Node: n,
		Messages: &queue.Messages{
			Values: make(map[int]struct{})},
		// Pending: make(map[string]*queue.Peer), // Not clear to me that initializing an empty queue of peers prior to receiving them is smart or worthwhile.
	}
}
