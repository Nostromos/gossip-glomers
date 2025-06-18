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
	ID   int    `json:"id"`
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
	Messages string `json:"messages"`
}

/*
{
  "type": "topology",
  "topology": {
    "n1": ["n2", "n3"],
    "n2": ["n1"],
    "n3": ["n1"]
  }
}
*/

type Topology map[string][]string

type TopologyReq struct {
	Type     string   `json:"type"` // "topology"
	Topology Topology `json:"topology"`
}

type TopologyOK struct {
	Type string `json:"type"` // "topology_ok"
}
