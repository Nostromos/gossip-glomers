package main

import (
	// --- Standard Lib ---
	"log"

	// --- Internal Lib ---
	"maelstrom-broadcast/internal/gossip"

	// --- Third Party ---
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	s := gossip.NewServer(n)

	n.Handle("echo", s.HandleEcho)
	n.Handle("generate", s.HandleGenerate)
	n.Handle("broadcast", s.HandleBroadcast)
	n.Handle("read", s.HandleRead)
	n.Handle("topology", s.HandleTopology)
	n.Handle("delta", s.HandleDelta)
	n.Handle("delta_ok", s.HandleDeltaOK)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
