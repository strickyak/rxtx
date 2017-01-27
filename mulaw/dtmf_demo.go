// +build main

/*
   Usage:    a.out 1-212-736-5000 >/dev/audio
*/
package main

import . "github.com/strickyak/rxtx/mulaw"

import (
	"math"
	"os"
	"strconv"
)

const RATE = 8000
const GAIN = 1000

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil { panic(err) }
	return f
}

func toneSample(freq, gain float64, t int) int16 {
	w := float64(t) * float64(freq) / float64(RATE) * 2.0 * math.Pi
	return int16(gain * math.Sin(w))
}

func dtmf(r rune) (float64, float64) {
	switch (r) {
	case '1': return 679, 1209
	case '2': return 679, 1336
	case '3': return 679, 1477
	case 'a': return 679, 1633
	case 'A': return 679, 1633
	case '4': return 770, 1209
	case '5': return 770, 1336
	case '6': return 770, 1477
	case 'b': return 770, 1633
	case 'B': return 770, 1633
	case '7': return 852, 1209
	case '8': return 852, 1336
	case '9': return 852, 1477
	case 'c': return 852, 1633
	case 'C': return 852, 1633
	case '*': return 941, 1209
	case '0': return 941, 1336
	case '#': return 941, 1477
	case 'd': return 941, 1633
	case 'D': return 941, 1633
	}
	return 0, 0
}

func main() {
	digits := os.Args[1]       // e.g. "411"
	buf := make([]byte, 1)

	for _, digit := range digits {
		for t := 0; t < int(0.5*RATE); t++ {
			f1, f2 := dtmf(digit)
			x1 := toneSample(f1, GAIN, t)
			x2 := toneSample(f2, GAIN, t)
			b := EncodeMulaw16(x1 + x2)
			buf[0] = b
			n, err := os.Stdout.Write(buf)
			if err != nil { panic(err) }
			if n != 1 { panic(n) }
		}
		for t := 0; t < int(0.5*RATE); t++ {
			buf[0] = 0
			n, err := os.Stdout.Write(buf)
			if err != nil { panic(err) }
			if n != 1 { panic(n) }
		}
	}
}
