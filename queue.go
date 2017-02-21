package rxtx

import (
	"fmt"
)

// Fixed size PacketQueue.  If you add too many,
// it drops the oldest ones.
const QLEN = PacketsPerSecond

type PacketQueue struct {
	Vec   []*Packet
	Begin int
	End   int
	Size  int
}

func NewPacketQueue() *PacketQueue {
	return &PacketQueue{Vec: make([]*Packet, QLEN)}
}

func (q PacketQueue) String() string {
	return fmt.Sprintf("%#v", q)
}
func (q *PacketQueue) Add(packet *Packet) {
	q.Begin = (q.Begin + 1) % QLEN
	q.Vec[q.Begin] = packet
	if q.Size == QLEN {
		// Drops the packet at the End.
		q.End = (q.End + 1) % QLEN
	} else {
		q.Size += 1
	}
}

func (q *PacketQueue) Take() *Packet {
	if q.Size == 0 {
		return nil
	} else {
		q.Size--
		q.End = (q.End + 1) % QLEN
		return q.Vec[q.End]
	}
}
