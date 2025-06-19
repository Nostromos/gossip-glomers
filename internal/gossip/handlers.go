package gossip

import (
	// --- Standard Lib
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
	msgValue := body["message"]
	num, ok := msgValue.(float64)
	if !ok {
		return fmt.Errorf("message is not a number")
	}
	if s.Messages.Add(int(num)) {
		for _, pq := range s.Pending {
			pq.Push(int(num))
		}
	}
	resp := map[string]any{
		"type": "broadcast_ok",
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleRead(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	resp := map[string]any{
		"type":     "read_ok",
		"messages": s.Messages.Snapshot(),
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleTopology(msg maelstrom.Message) error {
	var req map[string]any
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
	resp := map[string]any{
		"type":        "topology_ok",
		"in_reply_to": req["msg_id"],
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleDelta(msg maelstrom.Message) error {
	var req struct {
		Messages []int `json:"messages"`
	}
	_ = json.Unmarshal(msg.Body, &req)
	for _, v := range req.Messages {
		if s.Messages.Add(v) {
			for _, pq := range s.Pending {
				pq.Push(v)
			}
		}
	}
	resp := map[string]any{
		"type": "delta_ok",
	}
	return s.Node.Reply(msg, resp)
}

func (s *Server) HandleDeltaOK(msg maelstrom.Message) error {
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
