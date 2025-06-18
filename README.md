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
2. Run the binary:
  ```sh
  ./gossip-glomers
  ```

  OR run the project directly without building a binary:
  ```sh
  go run .
  ```