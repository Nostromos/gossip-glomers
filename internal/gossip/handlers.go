// Package gossip provides message handlers for the distributed gossip protocol.
// This file implements handlers for all Maelstrom message types including echo,
// unique ID generation, broadcast, read, topology setup, and delta synchronization.
package gossip

import (
	// --- Standard Lib ---
	"encoding/json"
	"fmt"

	// --- Internal Lib ---
	"maelstrom-broadcast/internal/queue"
	"maelstrom-broadcast/internal/protocol"

	// --- Third Party ---
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

// handle is a generic helper function that unmarshals a Maelstrom message body
// into the specified type T and executes the provided handler function.
// This eliminates boilerplate unmarshaling code across all message handlers.
//
// Parameters:
//   - msg: The incoming Maelstrom message
//   - fn: Handler function that processes the unmarshaled message
//
// Returns an error if unmarshaling fails or if the handler function returns an error.
func handle[T any](msg maelstrom.Message, fn func(T) error) error {
	var req T
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	return fn(req)
}

// HandleEcho processes echo requests from Maelstrom for connectivity testing.
// This is a simple ping-pong handler that immediately responds with an echo_ok message.
func (s *Server) HandleEcho(msg maelstrom.Message) error {
	return handle(msg, func(req protocol.EchoReq) error {
		resp := protocol.EchoOK{Type: "echo_ok"}
		return s.Node.Reply(msg, resp)	
	})
}

// HandleGenerate creates globally unique IDs for the distributed system.
// Uses atomic counter combined with node ID to ensure uniqueness across all nodes.
func (s *Server) HandleGenerate(msg maelstrom.Message) error {
	return handle(msg, func(req protocol.GenerateReq) error {
		uid := fmt.Sprintf("%s_%d", s.Node.ID(), s.Counter.Add(1))
		resp := protocol.GenerateOK{
			Type: "generate_ok",
			ID: uid, 
		}
	return s.Node.Reply(msg, resp)
	})
}

// HandleBroadcast receives new messages to be distributed across the network.
// If the message is new (not already seen), it's added to the global message set
// and queued for gossip propagation to all peer nodes.
func (s *Server) HandleBroadcast(msg maelstrom.Message) error {
	return handle(msg, func(req protocol.BroadcastReq) error {
		if s.Messages.Add(req.Message) {
			for _, pq := range s.Pending {
				pq.Add(req.Message)
			}
		}
		resp := protocol.BroadcastOK{
			Type: "broadcast_ok",
		}
		return s.Node.Reply(msg, resp)
	})
}

// HandleRead returns all messages currently known to this node.
// Provides a consistent snapshot of the distributed message set.
func (s *Server) HandleRead(msg maelstrom.Message) error {
	return handle(msg, func(req protocol.ReadReq) error {
		resp := protocol.ReadOK{
			Type:     "read_ok",
			Messages: s.Messages.GetSlice(),
		}
		return s.Node.Reply(msg, resp)
	})
}

// HandleTopology initializes the gossip network topology.
// Sets up peer queues for all other nodes and starts the background gossip loop.
// Uses sync.Once to ensure initialization happens only once per server instance.
func (s *Server) HandleTopology(msg maelstrom.Message) error {
	return handle(msg, func(req protocol.TopologyReq) error {
		s.initOnce.Do(func() {
			s.Pending = make(map[string]*queue.Peer)
			for _, peer := range s.Node.NodeIDs() {
				if peer == s.Node.ID() {
					continue
				}
				s.Pending[peer] = queue.NewPeerQueue()
			}
			go s.HandlePeerQueues()
		})
		resp := protocol.TopologyOK{
			Type:        "topology_ok",
		}
		return s.Node.Reply(msg, resp)
	})
}

// HandleDelta processes batch message updates from peer nodes in the gossip protocol.
// For each new message received, adds it to the local message set and propagates
// it to all other peers. This enables efficient gossip-based message dissemination.
func (s *Server) HandleDelta(msg maelstrom.Message) error {
	return handle(msg, func(req protocol.DeltaReq) error {
		for _, v := range req.Messages {
			if s.Messages.Add(v) {
				for _, pq := range s.Pending {
					pq.Add(v)
				}
			}
		}
		resp := protocol.DeltaOK{
			Type: "delta_ok",
		}
		return s.Node.Reply(msg, resp)
	})
}

// HandleDeltaOK processes acknowledgments from peers for successfully delivered delta messages.
// Stops the retry timer for the acknowledging peer to prevent unnecessary retransmissions.
// This implements the reliability mechanism for the gossip protocol.
// TODO: Much of this logic probably needs to live in PeerQueue and this needs proper typing with the DeltaOK struct.
func (s *Server) HandleDeltaOK(msg maelstrom.Message) error {
	return handle(msg, func(req protocol.DeltaOK) error {
		peerID := msg.Src // Maelstrom sets the sender ID here
		if pq, ok := s.Pending[peerID]; ok {
			pq.Clear()
		}
		return nil
	})
}
