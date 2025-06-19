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

func (s *Server) HandleEcho(msg maelstrom.Message) error {
	req := protocol.EchoReq{}
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	resp := protocol.EchoOK{Type: "echo_ok"}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleGenerate(msg maelstrom.Message) error {
	req := protocol.GenerateReq{}
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	idNum := s.Counter.Add(1)
	uid := fmt.Sprintf("%s_%d", s.Node.ID(), idNum)
	resp := protocol.GenerateOK{
		Type: "generate_ok",
		ID: uid, 
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleBroadcast(msg maelstrom.Message) error {
	req := protocol.BroadcastReq{}
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	if s.Messages.Add(req.Message) {
		for _, pq := range s.Pending {
			pq.Push(req.Message)
		}
	}
	resp := protocol.BroadcastOK{
		Type: "broadcast_ok",
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleRead(msg maelstrom.Message) error {
	req := protocol.ReadReq{}
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	resp := protocol.ReadOK{
		Type:     "read_ok",
		Messages: s.Messages.Snapshot(),
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleTopology(msg maelstrom.Message) error {
	req := protocol.TopologyReq{}
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	s.initOnce.Do(func() {
		s.Pending = make(map[string]*queue.Peer)
		for _, peer := range s.Node.NodeIDs() {
			if peer == s.Node.ID() {
				continue
			}
			s.Pending[peer] = &queue.Peer{
				Values: make(map[int]struct{}),
			}
		}
		go queue.HandlePeerQueues(s.Node, s.Pending)
	})
	resp := protocol.TopologyOK{
		Type:        "topology_ok",
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleDelta(msg maelstrom.Message) error {
	req := protocol.DeltaReq{}
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	for _, v := range req.Messages {
		if s.Messages.Add(v) {
			for _, pq := range s.Pending {
				pq.Push(v)
			}
		}
	}
	resp := protocol.DeltaOK{
		Type: "delta_ok",
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleDeltaOK(msg maelstrom.Message) error { // TODO: Much of this logic probably needs to live in PeerQueue and this needs proper typing with the DeltaOK struct.
	peerID := msg.Src // Maelstrom sets the sender ID here
		pq, ok := s.Pending[peerID]
		if !ok {
			return nil // unknown peer – ignore
		}
		pq.MU.Lock()
		if pq.Timer != nil { // delivery confirmed
			pq.Timer.Stop() // stop retry loop
			pq.Timer = nil
		}
		pq.MU.Unlock()
		return nil
}
