package rxtx

import "testing"

func makeVec(i int) []byte {
	return []byte{byte(i)}
}

func TestSegmentQueue(t *testing.T) {
	var i int
	var q = NewSegmentQueue()

	// Half fill the queue.
	for i = 0; i < QLEN/2; i++ {
		q.Add(makeVec(i))
	}
	if q.Size != QLEN/2 {
		t.Error(q.Size, QLEN)
	}
	// Then empty it.
	for i = 0; i < QLEN/2; i++ {
		z := q.Take()
		if z[0] != byte(i) {
			t.Error(i, z[0])
		}
	}
	if q.Size != 0 {
		t.Error(q.Size, QLEN)
	}

	// Fill the queue and overflow by 5.
	for i = 0; i < QLEN + 5; i++ {
		q.Add(makeVec(i))
	}
	if q.Size != QLEN {
		t.Error(q.Size, QLEN)
	}
	// 0..4 were lost.
	for i = 5; i < QLEN + 5; i++ {
		z := q.Take()
		if z[0] != byte(i) {
			t.Error(i, z[0])
		}
	}
	if q.Size != 0 {
		t.Error(q.Size, QLEN)
	}
}
