# Progress
A running log of progress and changes I'm making to my code to solve these problems but ALSO to better understand distributed systems.


## Refactor #1
### Why
`main.go` was getting long and difficult to understand, with lots of interwoven functionality. This made it tricky to isolate problems and led to a lot of jumping around. Additionally, it led to poor division of responsibility between functions/services/etc.

### How
Here's the structure I'm thinking of with help from o3:

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

All seems pretty straightforward for now. We have 3 areas of responsibility:
1. `package queue` - only deals with `messages` (a list of values a node has seen) and `pending` (a list of new values to send to peers) queues.
2. `package protocol` - JSON message types to ensure we're parsing and passing data easily and correctly.
3. `package gossip` - Handlers, node-specific methods, and logic for retries and backing-off.


| Date         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 06/18        | - project was restructured <br> - types mostly complete for handler logic <br>                                                                                                                                                                                                                                                                                                                                                                        |
| 06/19 midday | Completed types and handlers. Built server closure that wraps a node, our queues, and a few other things. Hanlders are methods on the server which makes passing data and calling functions easy. Also easier to test and reason about. There are still clear areas of improvement, particularly in duplicated queue logic (`Drain()`, `DrainAndSend()`, `HandlePeerQueues()`). This results in poor performance, especially with network partitions. |