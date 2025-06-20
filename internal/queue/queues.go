package queue

import "time"

type Peer struct {
	intSet
	Timer *time.Timer // want to store pointer to timer so if messages hang or drop we can resend
}

func NewPeerQueue() *Peer {
	return &Peer{
		intSet: newIntSet(),
	}
}

type Messages struct {
	intSet
}

func NewMessagesQueue() *Messages {
	return &Messages{
		intSet: newIntSet(),
	}
}