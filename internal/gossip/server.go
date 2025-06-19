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
	Messages *queue.Safe
	Pending  map[string]*queue.Peer
	Counter  atomic.Uint64
	initOnce sync.Once
}

func NewServer(n *maelstrom.Node) *Server {
	return &Server{
		Node: n,
		Messages: &queue.Safe{
			Values: make(map[int]struct{})},
		// Pending: make(map[string]*queue.Peer),
		// `Counter` left out - automatically zero
		// Learned that the counter doesn't need to be explicitly invoked or initialized
		// and that the zero value for `atomic.Uint64` is already valid just be declaring
		// it on the struct.
	}
} 
