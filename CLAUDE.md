``# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of Fly.io's Gossip Glomers distributed systems challenges. The project implements various distributed systems protocols including echo, unique ID generation, broadcast, and gossip protocols using the Maelstrom testing framework.

## Development Commands

### Build and Run
```bash
# Build the project
go build -o ~/go/bin/maelstrom-broadcast cmd/main.go

# Run maelstrom
./maelstrom/maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100
```

### Testing with Maelstrom
The binary is designed to be passed to Maelstrom for testing distributed systems scenarios. Example Maelstrom commands can be found in the test result directories under `store/`.

## Architecture

### Core Components

**Main Entry Point (`cmd/main.go`)**
- Sets up Maelstrom node and gossip server
- Registers message handlers for different protocol types
- Handles: echo, generate, broadcast, read, topology, delta, delta_ok

**Gossip Server (`internal/gossip/`)**
- `server.go`: Core server struct that wraps Maelstrom node with message queues and atomic counter
- `handlers.go`: Protocol message handlers using generic `handle[T]()` function for type-safe JSON unmarshaling
- `retry.go`: Placeholder for retry and backoff logic (currently empty)

**Queue System (`internal/queue/`)**
- `intset.go`: Thread-safe integer set implementation using sync.RWMutex as foundation for all queue types
- `queues.go`: Queue wrapper structs (`Messages` and `Peer`) that embed `intSet` for DRY implementation

**Protocol Types (`internal/protocol/types.go`)**
- JSON struct definitions for all message types (requests and responses)
- Includes: Echo, Generate, Broadcast, Read, Topology, Delta protocols

### Key Design Patterns

**Composition Over Inheritance**
- Both `Messages` and `Peer` structs embed `intSet` for shared thread-safe integer set functionality
- Eliminates code duplication while maintaining type safety

**Generic Programming**
- `handle[T any]()` function provides type-safe message unmarshaling with JSON validation
- Reduces boilerplate code across all protocol handlers

**Message Flow**
1. Messages received by handlers in `gossip/handlers.go`
2. New messages stored in thread-safe `Messages` queue (global message storage)
3. Messages propagated to `Peer` queues for gossip dissemination
4. Background goroutine (`HandlePeerQueues`) periodically drains peer queues and sends delta messages
5. Delta acknowledgments clear in-flight messages and handle retry logic

**Concurrency**
- Thread-safe message storage using `sync.RWMutex` in `intSet`
- Atomic counter for unique ID generation (`Server.Counter`)
- Per-peer retry timers and in-flight message tracking
- Background ticker for periodic gossip message sending (configurable `GossipInterval`)

**State Management**
- `Server.Messages`: Global set of all seen messages (embedded `intSet`)
- `Server.Pending`: Map of peer IDs to their pending message queues (`Peer` structs)
- `Server.Counter`: Atomic counter for generating unique IDs
- `initOnce`: Ensures topology setup happens only once
- Configurable timing: `GossipInterval` (50ms) and `RetryTimeout` (100ms)

### Queue Implementation Details

**intSet (`internal/queue/intset.go`)**
- Core thread-safe integer set with `Values map[int]struct{}`
- Operations: `Add()`, `Has()`, `GetSlice()`, `Clear()`
- **Note**: `DrainBatch(limit int)` method is incomplete - implementation is commented out in `queues.go`

**Peer Queue (`internal/queue/queues.go`)**
- Embeds `intSet` for message storage
- Additional fields: `InFlight []int` for retry tracking
- `DrainBatch(limit int)` method for batch processing with size limits (implementation in progress)

**Messages Queue (`internal/queue/queues.go`)**
- Simple wrapper around `intSet` for global message storage
- Used by `Server.Messages` for storing all seen messages

## Module Structure

The project uses Go modules with module name `maelstrom-broadcast` and depends on the Maelstrom Go library for distributed systems testing framework integration.

## Current State

The codebase is in a partially refactored state:
- **Completed**: Generic handlers, composition-based queue design, configurable timing
- **In Progress**: Batch size limiting for network efficiency (DrainBatch implementation incomplete)
- **Pending**: Retry logic implementation, race condition fixes