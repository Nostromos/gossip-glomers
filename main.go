package main

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var counter int = 0 // Should probably be uint64 TODO: Why will 32bit systems bork on this

type SafeQueue struct {
	mu     sync.RWMutex
	values map[int]struct{}
}

func (q *SafeQueue) Add(v int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.values[v]; exists {
		return false
	}
	q.values[v] = struct{}{}
	return true
}

func (q *SafeQueue) Has(v int) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for i := range q.values {
		if i == v {
			return true
		}
	}

	return false
}

func (q *SafeQueue) Snapshot() []int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	out := make([]int, 0, len(q.values))
	for v := range q.values {
		out = append(out, v)
	}
	return out
}

func (q *SafeQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.values = make(map[int]struct{})
}

type PeerQueue struct {
	mu    sync.RWMutex
	queue map[int]struct{}
	timer *time.Timer // want to store pointer to timer so if messages hang or drop we can resend
}

func (pq *PeerQueue) Push(v int) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if _, exists := pq.queue[v]; exists {
		return false
	}

	pq.queue[v] = struct{}{}
	return true
}

func (pq *PeerQueue) PopOne() (int, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for v := range pq.queue {
		delete(pq.queue, v)
		return v, true
	}
	return 0, false
}

func (pq *PeerQueue) Drain() []int {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	copy := make([]int, 0, len(pq.queue))
	for k, _ := range pq.queue {
		copy = append(copy, k)
		delete(pq.queue, k)
	}

	return copy
}

// func (pq *PeerQueue) Clear() {
// 	pq.mu.Lock()
// 	defer pq.mu.Unlock()

// 	pq.queue = make(map[int]struct{})
// }

func handlePeerQueues(node *maelstrom.Node, pending map[string]*PeerQueue) {
	ticker := time.NewTicker(100 * time.Millisecond)

	for range ticker.C {
		for peerID, pq := range pending {
			drainAndSend(node, peerID, pq)
		}
	}
}

func drainAndSend(node *maelstrom.Node, peer string, pq *PeerQueue) {
	batch := pq.Drain()

	if len(batch) == 0 {
		return
	}

	body := map[string]any{
		"type":     "delta",
		"messages": batch,
	}

	node.Send(peer, body)

	pq.mu.Lock()
	if pq.timer != nil {
		pq.timer.Stop()
	}

	pq.timer = time.AfterFunc(500 * time.Millisecond, func() {
		pq.mu.Lock()
		for _, v := range batch {
			pq.queue[v] = struct{}{}
		}

		pq.timer = nil
		pq.mu.Unlock()
	})

	pq.mu.Unlock()
}

func main() {
	n := maelstrom.NewNode()

	messages := &SafeQueue{
		values: make(map[int]struct{}),
	}

	var (
		pending   map[string]*PeerQueue
		startOnce sync.Once
	)

	// go handlePeerQueues(n, pending)

	n.Handle("echo", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "echo_ok"

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	n.Handle("generate", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "generate_ok"

		nodeID := n.ID()
		body["id"] = string(nodeID) + "_" + strconv.Itoa(counter)
		counter += 1

		// Echo the original message back with the updated message type.
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
			pending = make(map[string]*PeerQueue)

			for _, peer := range n.NodeIDs() {
				if peer == n.ID() {
					continue
				}

				pending[peer] = &PeerQueue{
					queue: make(map[int]struct{}),
				}
			}

			go handlePeerQueues(n, pending)
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
		peerID := msg.Src          // Maelstrom sets the sender ID here
    pq, ok := pending[peerID]
    if !ok {
        return nil               // unknown peer – ignore
    }

    pq.mu.Lock()        // delivery confirmed
    if pq.timer != nil {
        pq.timer.Stop()          // stop retry loop
        pq.timer = nil
    }
    pq.mu.Unlock()
    return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
