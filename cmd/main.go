package main

import (
	// --- Standard Lib
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	// --- Internal Lib ---
	"maelstrom-broadcast/internal/queue"

	// --- Third Party ---
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var counter int = 0 // Should probably be uint64 TODO: Why will 32bit systems bork on this

func main() {
	n := maelstrom.NewNode()
	
	messages := &queue.Safe{
		Values: make(map[int]struct{}),
	}

	var (
		pending   map[string]*queue.Peer
		startOnce sync.Once
	)

	n.Handle("echo", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "echo_ok"

		return n.Reply(msg, body)
	})

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "generate_ok"
		nodeID := n.ID()
		body["id"] = string(nodeID) + "_" + strconv.Itoa(counter)
		counter += 1

		return n.Reply(msg, body)
	})

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		msgValue := body["message"]
		num, ok := msgValue.(float64)
		if !ok {
			return fmt.Errorf("message is not a number")
		}
		if messages.Add(int(num)) {
			for _, pq := range pending {
				pq.Push(int(num))
			}
		}
		resp := map[string]any{
			"type": "broadcast_ok",
		}

		return n.Reply(msg, resp)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		resp := map[string]any{
			"type":     "read_ok",
			"messages": messages.Snapshot(),
		}

		return n.Reply(msg, resp)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var req map[string]any
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		startOnce.Do(func() {
			pending = make(map[string]*queue.Peer)
			for _, peer := range n.NodeIDs() {
				if peer == n.ID() {
					continue
				}
				pending[peer] = &queue.Peer{
					Values: make(map[int]struct{}),
				}
			}
			go queue.HandlePeerQueues(n, pending)
		})
		resp := map[string]any{
			"type":        "topology_ok",
			"in_reply_to": req["msg_id"],
		}

		return n.Reply(msg, resp)
	})

	n.Handle("delta", func(msg maelstrom.Message) error {
		var req struct {
			Messages []int `json:"messages"`
		}
		_ = json.Unmarshal(msg.Body, &req)
		for _, v := range req.Messages {
			if messages.Add(v) {
				for _, pq := range pending {
					pq.Push(v)
				}
			}
		}
		resp := map[string]any{
			"type": "delta_ok",
		}

		return n.Reply(msg, resp)
	})

	n.Handle("delta_ok", func(msg maelstrom.Message) error {
		peerID := msg.Src // Maelstrom sets the sender ID here
		pq, ok := pending[peerID]
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
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
