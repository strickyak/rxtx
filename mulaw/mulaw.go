package mulaw

import "log"

var boundary = []int16{4063, 2015, 991, 479, 223, 95, 31, 1}

// Decode one mulaw sample byte.
func DecodeMulaw16(b byte) int16 {
	if b == 0xFF {
		return 0
	}
	if b == 0x7F {
		return -1
	}
	sign := b & 0x80
	exp := (b & 0x70) >> 4
	man := b & 0x0F

	sz := 0x100 >> uint(exp)

	var x int16 = int16(sz)*int16(15-man) + boundary[exp]

	if exp == 7 {
		x -= 2
	}

	if sign == 0 { // if negative
		return ^x
	}
	return x
}

// Encode one mulaw sample byte.
func EncodeMulaw16(x int16) byte {
	var neg bool
	if x < 0 {
		neg = true
		x = ^x
	}
	// Clip.
	if x > 8158 {
		x = 8158
	}
	var sz int16 = 256
	var m int16 = 0
	var i int16
	for i = 0; i < 8; i++ {
		b := boundary[i]
		if x >= b {
			m = 15 - ((x - b) / sz)
			if i == 7 {
				m--
			}
			break
		}
		sz >>= 1
	}
	if m < 0 {
		log.Panicf("%d", m)
	}
	if m > 15 {
		log.Panicf("%d", m)
	}
	var z byte = byte((i << 4) | m)
	if x == 0 {
		z = 0x7F // Special case.
	}
	if z < 0 {
		log.Panicf("%d", z)
	}
	if z > 255 {
		log.Panicf("%d", z)
	}
	if !neg {
		z |= 0x80 // Positive case has sign bit.
	}
	return z
}
