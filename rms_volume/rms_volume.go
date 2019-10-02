// Inputs is raw s16_le audio.  Prints RMS volume for chunks.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

var rate = flag.Float64("r", 44100, "samples per second")
var duration = flag.Duration("d", 100*time.Millisecond, "duration of chunks")

var buf []byte

func Chunk() bool {
	bc, err := os.Stdin.Read(buf)
	n := bc / 2 // Two bytes per sample
	var sumsq float64
	for i := 0; i < n; i++ {
		lo := buf[i+i]
		hi := buf[i+i+1]
		var x int
		if (hi & 0x80) != 0 {
			// neg
			x = 0x8000 - (int(hi&0x7F) << 8) - int(lo)
		} else {
			// pos
			x = (int(hi) << 8) + int(lo)
		}
		sumsq += float64(x)
	}
	rms := math.Sqrt(sumsq / float64(n))

	var b bytes.Buffer
	for i := 0; i < int(rms); i++ {
		b.WriteRune('#')
	}
	fmt.Printf("%10.6f : %s\n", rms, b.String())
	return err == nil
}

func main() {
	flag.Parse()
	log.Printf("rate %v duration %v", rate, duration)
	secs := duration.Seconds()
	n := int(*rate * secs)
	log.Printf("secs %v n %v", secs, n)
	buf = make([]byte, n)
	for Chunk() {
	}
	//fmt.Printf("\n")
}
