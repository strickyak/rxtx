package rxtx

type BitSet byte

func Bit(x int) BitSet {
	Assert(0 <= x)
	Assert(x <= 7)
	return BitSet(1 << uint(x))
}

func (b BitSet) Has(f byte) bool {
	return (byte(b) & f) == f
}
