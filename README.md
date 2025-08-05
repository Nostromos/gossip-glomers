<p align="center">
  <img src="https://fly.io/blog/gossip-glomers/assets/gossip-glomers.webp" alt="Gossip Glomers by Fly.io" />
</p>
<h1 align="center"><i>Gossip Glomers</i></h1>

<p align="center">
  <a>
    <img alt="Go" src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  </a>
  <a>
    <img alt="Distributed Systems" src="https://img.shields.io/badge/Distributed%20Systems-FF6B6B?style=for-the-badge&logo=apache&logoColor=white" />
  </a>
  <a>
    <img alt="Maelstrom" src="https://img.shields.io/badge/Maelstrom-4A5568?style=for-the-badge&logo=testing-library&logoColor=white" />
  </a>
</p>

## Overview

This is my solution to [Fly.io's Gossip Glomers](https://fly.io/dist-sys/) distributed systems challenges. The project implements various distributed protocols including echo, unique ID generation, broadcast, and gossip protocols using the Maelstrom testing framework.

## Getting Started

### Prerequisites
- Go (version 1.21+ recommended)
- Git
- [Maelstrom](https://github.com/jepsen-io/maelstrom) testing framework

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Nostromos/gossip-glomers
   cd gossip-glomers
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Download Maelstrom (if not already installed):
   ```bash
   wget https://github.com/jepsen-io/maelstrom/releases/download/v0.2.3/maelstrom.tar.bz2
   tar -xjf maelstrom.tar.bz2
   ```

## Build and Run

### Building the Project

```bash
go build -o ~/go/bin/maelstrom-broadcast cmd/main.go
```

### Running with Maelstrom

Test various challenges:

```bash
# Echo protocol
./maelstrom/maelstrom/maelstrom test -w echo --bin ~/go/bin/maelstrom-broadcast --node-count 1 --time-limit 10

# Unique ID generation
./maelstrom/maelstrom/maelstrom test -w unique-ids --bin ~/go/bin/maelstrom-broadcast --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

# Broadcast (single-node)
./maelstrom/maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 1 --time-limit 20 --rate 10

# Broadcast (multi-node)
./maelstrom/maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10

# Broadcast with network partitions
./maelstrom/maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition

# Stress test
./maelstrom/maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100
```

## Project Structure

```
gossip-glomers/
├── cmd/
│   └── main.go              # Main entry point
├── internal/
│   ├── gossip/              # Core gossip protocol implementation
│   │   ├── server.go        # Server struct and initialization
│   │   ├── handlers.go      # Message handlers for different protocols
│   │   └── retry.go         # Retry logic (WIP)
│   ├── protocol/            # Protocol message definitions
│   │   └── types.go         # JSON struct definitions for all message types
│   └── queue/               # Thread-safe queue implementations
│       ├── intset.go        # Base thread-safe integer set
│       └── queues.go        # Message and peer queue implementations
├── store/                   # Maelstrom test results and logs
├── CLAUDE.md               # AI assistant instructions
└── README.md               # This file
```

## Architecture

### Core Components

- **Server** (`internal/gossip/server.go`): Main server that wraps Maelstrom node with message queues and atomic counter
- **Handlers** (`internal/gossip/handlers.go`): Protocol-specific message handlers using generic type-safe unmarshaling
- **Queue System** (`internal/queue/`): Thread-safe data structures for message storage and peer communication
- **Protocol Types** (`internal/protocol/types.go`): Type definitions for all protocol messages

### Implemented Protocols

- **Echo**: Simple echo service
- **Generate**: Unique ID generation using node ID + atomic counter
- **Broadcast**: Message broadcast with gossip propagation
- **Read**: Query for all known messages
- **Topology**: Network topology configuration
- **Delta**: Gossip protocol for efficient message synchronization

### Key Design Features

- Thread-safe operations using `sync.RWMutex`
- Generic message handling with `handle[T]()` function
- Composition-based design (queue types embed `intSet`)
- Configurable timing parameters (50ms gossip interval, 100ms retry timeout)
- Background goroutines for periodic message propagation

### Message Flow

1. Messages received by handlers in `gossip/handlers.go`
2. New messages stored in thread-safe `Messages` queue
3. Messages queued for propagation to peer nodes
4. Background goroutine periodically sends delta messages
5. Delta acknowledgments clear in-flight messages

## Development

### Current Status

**Completed:**
- Basic protocol implementations (echo, generate, broadcast, read)
- Gossip-based message propagation
- Thread-safe message storage
- Delta synchronization protocol

**In Progress:**
- Batch size limiting (partially implemented)
- Retry logic for failed messages

**Not Started:**
- Advanced fault tolerance
- Performance optimizations for challenges #3d and #3e
- Grow-only counter (Challenge #4)
- Kafka-style log (Challenge #5)
- Transactions (Challenge #6)

### Testing

Run tests and view results:

```bash
# Run test
./maelstrom/maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20

# View results
open store/latest/index.html
```

## Notes

- Don't run the binary alone - pass it to Maelstrom for testing.
- Test results are stored in the `store/` directory with detailed performance metrics
- See `CLAUDE.md` for detailed architectural documentation

## References

- [Fly.io Gossip Glomers Challenges](https://fly.io/dist-sys/)
- [Maelstrom Documentation](https://github.com/jepsen-io/maelstrom/blob/main/doc/01-getting-ready/index.md)
- [Distributed Systems Course](https://www.distributedsystemscourse.com/)

## License

See the [LICENSE](./LICENSE) file for details.