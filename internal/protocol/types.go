// Package protocol defines JSON message types for Maelstrom distributed systems testing.
// Implements request/response pairs for echo, unique ID generation, broadcast,
// topology management, and gossip delta synchronization protocols.
package protocol

// EchoReq represents an echo request message for connectivity testing.
// Simple ping-pong protocol to verify message routing and basic communication
// between nodes in the distributed system.
type EchoReq struct {
	Type  string `json:"type"` // "echo"
	MsgID int    `json:"msg_id"`
	Echo  string `json:"echo"`
}

// EchoOK represents the response to an echo request.
// Returns the same echo payload to confirm successful message processing
// and round-trip communication capability.
type EchoOK struct {
	Type      string `json:"type"` // "echo_ok"
	MsgID     int    `json:"msg_id"`
	InReplyTo int    `json:"in_reply_to"`
	Echo      string `json:"echo"`
}

// GenerateReq represents a request for globally unique ID generation.
// Used to test distributed unique identifier creation across multiple nodes
// without coordination or central authority.
type GenerateReq struct {
	Type string `json:"type"` // "generate"
}

// GenerateOK represents the response containing a globally unique identifier.
// The ID field contains a string guaranteed to be unique across all nodes
// in the distributed system, typically combining node ID with local counter.
type GenerateOK struct {
	Type string `json:"type"` // "generate_ok"
	ID   string `json:"id"`
}

// BroadcastReq represents a message broadcast request to all nodes.
// Contains an integer message that should be propagated throughout
// the distributed system using gossip protocols for eventual consistency.
type BroadcastReq struct {
	Type    string `json:"type"` // "broadcast"
	MsgID   int    `json:"msg_id"`
	Message int    `json:"message"`
}

// BroadcastOK represents acknowledgment of a broadcast message.
// Confirms that the node has received and processed the broadcast request,
// though it doesn't guarantee propagation to other nodes.
type BroadcastOK struct {
	Type      string `json:"type"` // "broadcast_ok"
	InReplyTo int    `json:"in_reply_to"`
}

// ReadReq represents a request to read all messages seen by this node.
// Used to query the current state of broadcast messages for verification
// and testing of eventual consistency in the gossip protocol.
type ReadReq struct {
	Type string `json:"type"` // "read"
}

// ReadOK represents the response containing all known messages.
// Messages field contains a slice of all integer messages this node
// has seen, either directly or through gossip propagation.
type ReadOK struct {
	Type     string `json:"type"` // "read_ok"
	Messages []int  `json:"messages"`
}

// Topology represents the network topology as a map of node connections.
// Maps each node ID to a slice of its directly connected neighbor node IDs,
// defining the communication graph for gossip message propagation.
type Topology map[string][]string

// TopologyReq represents a topology configuration message.
// Sent by Maelstrom to inform nodes about their network neighbors
// and establish the communication topology for testing scenarios.
type TopologyReq struct {
	Type     string   `json:"type"` // "topology"
	Topology Topology `json:"topology"`
}

// TopologyOK represents acknowledgment of topology configuration.
// Confirms that the node has received and processed the topology
// information and is ready for distributed protocol testing.
type TopologyOK struct {
	Type string `json:"type"` // "topology_ok"
}

// DeltaReq represents a gossip delta synchronization message.
// Contains a batch of message IDs being shared between peer nodes
// for efficient propagation and eventual consistency achievement.
type DeltaReq struct {
	Type     string `json:"type"` // "delta"
	Messages []int  `json:"messages"`
}

// DeltaOK represents acknowledgment of a delta synchronization message.
// Confirms successful receipt of gossip messages and triggers cleanup
// of in-flight message tracking for retry logic management.
type DeltaOK struct {
	Type string `json:"type"` // "delta_ok"
}