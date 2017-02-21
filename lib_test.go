package rxtx

import "testing"

func makePacket(i int) *Packet {
	return &Packet{nil, []byte{byte(i)}, nil}
}

func TestSegmentQueue(t *testing.T) {
	var i int
	var q = NewPacketQueue()

	// Half fill the queue.
	for i = 0; i < QLEN/2; i++ {
		q.Add(makePacket(i))
	}
	if q.Size != QLEN/2 {
		t.Error(q.Size, QLEN)
	}
	// Then empty it.
	for i = 0; i < QLEN/2; i++ {
		z := q.Take()
		if z.Segment[0] != byte(i) {
			t.Error(i, z.Segment[0])
		}
	}
	if q.Size != 0 {
		t.Error(q.Size, QLEN)
	}

	// Fill the queue and overflow by 5.
	for i = 0; i < QLEN+5; i++ {
		q.Add(makePacket(i))
	}
	if q.Size != QLEN {
		t.Error(q.Size, QLEN)
	}
	// 0..4 were lost.
	for i = 5; i < QLEN+5; i++ {
		z := q.Take()
		if z.Segment[0] != byte(i) {
			t.Error(i, z.Segment[0])
		}
	}
	if q.Size != 0 {
		t.Error(q.Size, QLEN)
	}
}
