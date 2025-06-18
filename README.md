# Gossip Glomers

This is my solution to [Fly.io's](https://fly.io) set of distributed systems challenges, Gossip Glomers.

## Getting Started

### Prerequisites
- Go (version 1.21+ recommended)
- Git
- Maelstrom

### Installation

1. Clone the repo and go to the folder:
  ```sh
  git clone https://github.com/Nostromos/gossip-glomers && cd gossip-glomers
  ```
2. Install dependencies:
  ```sh
  go mod tidy
  ```

### Build and Run

1. Build the project:
  ```sh
  go build -o gossip-glomers .
  ```

> [!WARNING]
> You *should* be passing the binary to maelstrom and generally shouldn't be running it alone, but I'm including the instructions for completeness.

2. Run the binary:
  ```sh
  ./gossip-glomers
  ```

  OR run the project directly without building a binary:
  ```sh
  go run .
  ```

## Refactor

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
│   ├── protocol/
│   │   └── types.go         // define structs with JSON info per message type
│   ├── docs/
│   │   └── REFACTOR.md      // detail refactor approach and document how it goes/went
│   └── README.md        
├── go.mod
└── go.sum
```

All seems pretty straightforward for now.