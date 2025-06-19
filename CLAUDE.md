`# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of Fly.io's Gossip Glomers distributed systems challenges. The project implements various distributed systems protocols including echo, unique ID generation, broadcast, and gossip protocols using the Maelstrom testing framework.

## Development Commands

### Build and Run
```bash
# Build the project
go build -o gossip-glomers .

# Run directly without building binary
go run .

# Install/update dependencies
go mod tidy
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
- `handlers.go`: Protocol message handlers for all supported operations
- `retry.go`: Retry and backoff logic (if present)

**Queue System (`internal/queue/`)**
- `safe_queue.go`: Thread-safe message storage using sync.RWMutex
- `peer_queue.go`: Per-peer message queues with retry timers for gossip protocol

**Protocol Types (`internal/protocol/types.go`)**
- JSON struct definitions for all message types (requests and responses)
- Includes: Echo, Generate, Broadcast, Read, Topology, Delta protocols

### Key Design Patterns

**Message Flow**
1. Messages received by handlers in `gossip/handlers.go`
2. New messages stored in thread-safe `queue.Safe` 
3. Messages propagated to peer queues for gossip dissemination
4. Background goroutine (`HandlePeerQueues`) periodically drains peer queues and sends delta messages
5. Retry mechanism ensures delivery with exponential backoff

**Concurrency**
- Thread-safe message storage using `sync.RWMutex`
- Atomic counter for unique ID generation
- Per-peer retry timers for handling network partitions
- Background ticker for periodic gossip message sending

**State Management**
- `Server.Messages`: Global set of all seen messages
- `Server.Pending`: Map of peer IDs to their pending message queues
- `Server.Counter`: Atomic counter for generating unique IDs
- `initOnce`: Ensures topology setup happens only once

## Module Structure

The project uses Go modules with module name `maelstrom-broadcast` and depends on the Maelstrom Go library for distributed systems testing framework integration.