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

## PeerQueue Refactor

*From Claude...*:

>Key Issues & Refactoring Suggestions
>
>  1. Naming consistency: Peer struct should be PeerQueue for clarity
>  2. Method naming: PopOne() and Drain() have unclear semantics - consider TryPop() and DrainAll()
>  3. Race condition: Lines 79-87 have a potential race condition with nested mutex operations
>  4. Timer management: Complex timer logic could be extracted into separate methods
>  5. Magic numbers: Hard-coded 100 * time.Millisecond should be configurable
>  6. Error handling: No error handling for node.Send() failures
>  7. Memory efficiency: Drain() creates unnecessary slice copies
>
>  Specific Improvements
>
>  Structure & Naming:
>  - Rename Peer → PeerQueue
>  - Add constructor function NewPeerQueue()
>  - Extract timer duration as configurable constant
>
>  Concurrency Safety:
>  - Fix race condition in timer callback
>  - Simplify mutex usage patterns
>  - Consider using channels for timer coordination
>
>  Performance:
>  - Optimize Drain() to avoid double iteration
>  - Add batch size limits for large queues
>  - Consider using sync.Pool for slice reuse
>
>  Maintainability:
>  - Extract retry logic into separate RetryManager
>  - Add proper error handling and logging
>  - Split drainAndSend into smaller functions

*From ChatGPT:*

queue package
Keep it small and dumb – just concurrent‐safe containers:
| file | purpose |
| ---- | ------- |
| safe.go | Safe set for Messages |
| peer.go (was peer_queue.go) | Peer ⇒ holds its per-peer Values set plus the retry Timer & mutex |

There is no global “peer-queue” type any more; the map on Server
(`map[string]*queue.Peer`) is the peer-queue collection.
	•	gossip.Server
Owns lifecycle things, so flushPeers belongs here, not in queue.

	•	Keep queue for tiny, concurrency-focused data structures (Safe, Peer).
	•	Own all lifecycle/background behaviour inside gossip.Server
(flushPeers, retry timers).
	•	Launch the flusher once you process the first topology; no background polling
of NodeIDs() required.

	1.	Per-peer retry policy
In your current flushPeers() / HandlePeerQueues logic you:
	•	move a batch out of the Peer queue
	•	call node.Send(peer, batch)
	•	start a single 200 ms timer that re-enqueues the batch once.
If either the delta or the peer’s delta_ok is lost a second time, we give up.
→ Keep the entry in the queue until the ack arrives.  A simple way is to mark it “in-flight” but do not delete it; resend after 200 ms again until you finally see delta_ok.
	2.	Timer granularity
Tick frequency is 100 ms, retry 200 ms.
That means worst-case you wait ~300 ms between attempts.
With partitions and drops that’s fine, but 30 s / 0.3 s ≈ 100 attempts.
Yet we stopped after one.  Tightening the retry interval helps but the real fix is unlimited retries.
	3.	Starting the loop
Make sure you do not start flushPeers() until after the topology message is processed, otherwise every node spends the first few hundred milliseconds happily trying to talk to an empty peer map (wasted messages).
	4.	Batch size
You drain the entire set each tick.  That works, but if the message is big it is more likely to be dropped.  Consider smallish batches (< 64 ints).

⸻


### Todos

Structure & Naming:
  - [x] Rename Peer → PeerQueue
  - [x] Add constructor function NewPeerQueue()
  - [ ] Extract timer duration as configurable constant
  - [ ] Move peer queue to gossip.Server, along with things like flushPeers

  Concurrency Safety:
  - [ ] Fix race condition in timer callback
  - [ ] Simplify mutex usage patterns
  - [ ] Consider using channels for timer coordination

  Performance:
  - [ ] Optimize Drain() to avoid double iteration
  - [ ] Add batch size limits for large queues
  - [ ] Consider using sync.Pool for slice reuse

  Maintainability:
  - [ ] Extract retry logic into separate RetryManager
  - [ ] Add proper error handling and logging
  - [ ] Split drainAndSend into smaller functions

	•	Keep queue for tiny, concurrency-focused data structures (Safe, Peer).
	•	Own all lifecycle/background behaviour inside gossip.Server
(flushPeers, retry timers).
	•	Launch the flusher once you process the first topology; no background polling
of NodeIDs() required.

- [ ] Per-peer retry policy
        In your current flushPeers() / HandlePeerQueues logic you:
          •	move a batch out of the Peer queue
          •	call node.Send(peer, batch)
          •	start a single 200 ms timer that re-enqueues the batch once.
        If either the delta or the peer’s delta_ok is lost a second time, we give up.
- [ ] Keep the entry in the queue until the ack arrives.  A simple way is to mark it “in-flight” but do not delete it; resend after 200 ms again until you finally see delta_ok.
- [ ] Add unlimited retries
        Tick frequency is 100 ms, retry 200 ms.
        That means worst-case you wait ~300 ms between attempts.
        With partitions and drops that’s fine, but 30 s / 0.3 s ≈ 100 attempts.
        Yet we stopped after one.  Tightening the retry interval helps but the real fix is unlimited retries.
- [ ] Starting the loop
        Make sure you do not start flushPeers() until after the topology message is processed, otherwise every node spends the first few hundred milliseconds happily trying to talk to an empty peer map (wasted messages).
- [ ] Batch size
        You drain the entire set each tick.  That works, but if the message is big it is more likely to be dropped.  Consider smallish batches (< 64 ints).