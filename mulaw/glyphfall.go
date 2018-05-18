// +build main

/*
   Produce audio for printing ASCII onto radio waterfall.

   Uses 500, 600, 700, 800, and 900 Hz tones for five columns.

   Usage:    a.out 'Test DE W6REK' >/dev/audio

   go run glyphfall.go -p 3000 "Test DE W6REK" | pacat --format=mulaw --rate=8000 --channels=1 -d DEVICE

   (Use "pacmd list" to discover DEVICE names.)
*/
package main

import "github.com/strickyak/rxtx/mulaw"
import "github.com/strickyak/rxtx/font5x7"

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
)

var PixelTime = flag.Int("p", 800, "Rate. How many 1/8000 sec to spend on each pixel.")

var printf = fmt.Printf

const SAMPLE_RATE = 8000

func toneSample(freq, gain float64, t int) int16 {
	omega := float64(t) * float64(freq) / float64(SAMPLE_RATE) * 2.0 * math.Pi
	return int16(gain * math.Sin(omega))
}

const GAIN_STEP = 20.0 // Adjust gain slowly for output tones, to avoid splatter.

func adjust(gain float64, maxGain float64, tone byte) float64 {
	if tone == 0 {
		gain -= GAIN_STEP
		if gain < 0.0 {
			gain = 0.0
		}
	} else {
		gain += GAIN_STEP
		if gain > maxGain {
			gain = maxGain
		}
	}
	return gain
}

func render(a []byte) []byte {
	var z []byte
	var g1, g2, g3, g4, g5 float64
	t := 0
	for _, row := range a {
		for i := 0; i < *PixelTime; i++ {
			var x int16
			g1 = adjust(g1, 2000, (row & 0x10))
			g2 = adjust(g2, 1800, (row & 0x08))
			g3 = adjust(g3, 1600, (row & 0x04))
			g4 = adjust(g4, 1400, (row & 0x02))
			g5 = adjust(g5, 1200, (row & 0x01))
			x += toneSample(500.0, g1, t)
			x += toneSample(600.0, g2, t)
			x += toneSample(700.0, g3, t)
			x += toneSample(800.0, g4, t)
			x += toneSample(900.0, g5, t)
			audio := mulaw.EncodeMulaw16(x)
			z = append(z, audio)
			// println(t, x, audio, int(g1), int(g2), int(g3), int(g4), int(g5))
			t++
		}
	}
	return z
}

func main() {
	flag.Parse()

	message := strings.Join(flag.Args(), " ")
	bitmap := font5x7.VerticalStringFiveBitsWide(message)
	audio := render(bitmap)
	os.Stdout.Write(audio)
}
