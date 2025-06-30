# Refactor #2

```
Spec:

We will increase our node count to 25 and add a delay of 100ms to each message to simulate a slow network. This could be geographic latencies (such as US to Europe) or it could simply be a busy network.

Your challenge is to achieve the following:

- Messages-per-operation is below 30
- Median latency is below 400ms
- Maximum latency is below 600ms

Feel free to ignore the topology you’re given by Maelstrom and use your own; it’s only a suggestion. Don’t compromise safety under faults. Double-check that your solution is still correct (even though it will be much slower) with --nemesis partition
```

## Key Challenges

- Messages-per-operation < 30: Current implementation likely floods network
- Median latency < 400ms: Need smart topology to minimize hops
- Max latency < 600ms: Avoid worst-case scenarios

Strategic Approach

1. Topology Optimization
   - Replace current topology with efficient structure (tree, ring, or hybrid)
   - Minimize diameter while maintaining fault tolerance
   - Consider hub-and-spoke or spanning tree for 25 nodes

2. Batch Message Aggregation
   - Implement the incomplete DrainBatch() functionality
   - Send multiple messages in single delta operations
   - Reduce network chattiness significantly

3. Adaptive Timing
   - Tune GossipInterval and RetryTimeout for 100ms network delay
   - Balance between latency and message efficiency
   - Consider exponential backoff for retries

4. Smart Dissemination
   - Avoid broadcasting to all peers simultaneously
   - Use selective forwarding based on topology structure
   - Implement message deduplication at receiver level

The current codebase has good foundations with thread-safe queues and configurable timing - we mainly need to
optimize the topology strategy and complete the batching implementation.

## Implementation Steps

### Phase 1: Foundation (High Priority)

#### Step 1: Baseline Analysis
   - Run current implementation with 25 nodes, measure messages-per-operation and latencies
   - Identify which parts of code generate the most network traffic
   - Document current topology handling in handlers.go

#### Step 2: Complete Batching Infrastructure
   - Implement DrainBatch(limit int) method in internal/queue/queues.go
   - Add batch size configuration to server struct
   - Test batch functionality with unit tests

#### Step 3: Topology Design
   - Research optimal topologies for 25 nodes (likely spanning tree or ring)
   - Calculate theoretical diameter and message complexity
   - Design topology generation algorithm in new topology.go file

### Phase 2: Core Implementation (Medium Priority)

#### Step 4: Smart Routing
- Modify topology handler to ignore Maelstrom's topology
- Implement topology-aware peer selection in HandlePeerQueues()
- Add neighbor-only forwarding instead of broadcast-to-all

#### Step 5: Timing Optimization
- Adjust GossipInterval and RetryTimeout for 100ms network delay
- Add adaptive backoff for failed messages
- Test different timing configurations

#### Step 6: Batch Message Sending
- Integrate DrainBatch() into HandlePeerQueues()
- Limit delta messages to reasonable batch sizes (10-20 messages)
- Implement message aggregation before sending

### Phase 3: Validation (Low Priority)

#### Step 7: Performance Testing
- Test with Maelstrom using 25 nodes and 100ms latency
- Measure and optimize until meeting all three metrics
- Profile bottlenecks and iterate

#### Step 8: Fault Tolerance
- Validate with --nemesis partition flag
- Ensure correctness under network partitions
- Fix any safety issues discovered 

## Implementation Progress

### Baseline Analysis

I ran the current implementation with 25 nodes, latency at 100ms, and at a rate of 100. Here's where our north star metrics stand:

| Metric Name            | Metric Goal | Metric Value | Delta    |
| ---------------------- | ----------- | ------------ | -------- |
| Messages-per-operation | <30         | 3598.565     | +11,900% |
| Median Latency         | <400ms      | 77ms         | -81%     |
| Maximum Latency        | <600ms      | 711ms        | +18.5%   |

Messages-per-operation (MPO) is an astounding 119x our goal. We are sending a *lot* of messages. 

Median latency is great, probably because we don't make any hops, and every node is just screaming its messages into the void until other nodes respond (or don't).

Max latency isn't horrible either - I was expecting worse. My gut tells me that if we reduce MPO, there's much less processing to be done and it'll be easier for nodes to "hear" updates, thus reducing all latencies. As an aside, whatever I do I need to ensure I"m keeping message lifecycles sub-500ms at least - we will introduce 100ms latency by default to reflect a more real-world environment. 

$$