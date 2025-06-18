## Refactor #1

`main.go` is getting long and difficult to parse. Feels like its time for a refactor. Here's the structure I'm thinking of with help from o3:

```
gossip-glomers/
├── cmd/
│   └── main.go              // maelstrom node, wiring up handlers
├── internal/
│   ├── queue/
│   │   ├── safe_queue.go    // node messages
│   │   └── peer_queue.go    // list of messages per peer
│   ├── gossip/
│   │   ├── node.go          // topology, handlers
│   │   └── retry.go         // retries and backoffs logic
│   └── protocol/
│       └── types.go         // define structs with JSON info per message type
├── docs/
│   └── REFACTOR.md          // detail refactor approach and document how it goes/went
├── README.md        
├── go.mod
└── go.sum
```

All seems pretty straightforward for now.