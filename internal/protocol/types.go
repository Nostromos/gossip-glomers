package protocol

type EchoReq struct {
	Type  string `json:"type"` // "echo"
	MsgID int    `json:"msg_id"`
	Echo  string `json:"echo"`
}

type EchoOK struct {
	Type      string `json:"type"` // "echo_ok"
	MsgID     int    `json:"msg_id"`
	InReplyTo int    `json:"in_reply_to"`
	Echo      string `json:"echo"`
}

type GenerateReq struct {
	Type string `json:"type"` // "generate"
}

type GenerateOK struct {
	Type string `json:"type"` // "generate_ok"
	ID   string    `json:"id"`
}

type BroadcastReq struct {
	Type    string `json:"type"` // "broadcast"
	MsgID   int    `json:"msg_id"`
	Message int    `json:"message"`
}

type BroadcastOK struct {
	Type      string `json:"type"` // "broadcast_ok"
	InReplyTo int    `json:"in_reply_to"`
}

type ReadReq struct {
	Type string `json:"type"` // "read"
}

type ReadOK struct {
	Type     string `json:"type"` // "read_ok"
	Messages []int `json:"messages"`
}

type Topology map[string][]string

type TopologyReq struct {
	Type     string   `json:"type"` // "topology"
	Topology Topology `json:"topology"`
}

type TopologyOK struct {
	Type string `json:"type"` // "topology_ok"
}

type DeltaReq struct {
	Type string `json:"type"` // "delta"
	Messages []int `json:"messages"`
}

type DeltaOK struct {
	Type     string `json:"type"` // "delta_ok"
}