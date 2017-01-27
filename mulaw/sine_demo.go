// +build main

/*
    Usage:  a.out freq_hz duration_secs gain > /dev/audio
*/
package main

import . "github.com/strickyak/rxtx/mulaw"

import (
	"math"
	"os"
	"strconv"
)

const RATE = 8000

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil { panic(err) }
	return f
}

func main() {
	freq := parseFloat(os.Args[1])      // e.g.  500 Hz
	duration := parseFloat(os.Args[2])  // e.g.   30 sec
	gain := parseFloat(os.Args[3])      // e.g. 1000 gain
	buf := make([]byte, 1)

	for t := 0; t < int(duration*RATE); t++ {
		w := float64(t) * float64(freq) / float64(RATE) * 2.0 * math.Pi
		x := int16(gain * math.Sin(w))
		b := EncodeMulaw16(x)
		buf[0] = b
		n, err := os.Stdout.Write(buf)
		if err != nil { panic(err) }
		if n != 1 { panic(n) }
	}
}
