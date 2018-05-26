// +build main

// go run nested_dtcw_main.go | pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100 --device=alsa_output.usb-Burr-Brown_from_TI_USB_Audio_CODEC-00.analog-stereo
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
)
import . "github.com/strickyak/rxtx/qrss"

var RATE = flag.Float64("rate", 44100, "Audio Sample Rate")
var SECS = flag.Float64("secs", 6, "Tone length in secs")
var RAMP = flag.Float64("ramp", 1.0, "Ramp up/down time in secs")
var GAIN = flag.Float64("gain", 0.86, "Modulation Gain")
var BASE = flag.Float64("base", 500, "Base Hz")
var RAND = flag.Float64("base_rand", 100, "Random addition to Base Hz")
var STEP = flag.Float64("step", 4, "Tone Step Hz")

var MODE = flag.String("mode", "nested", "nested | ")

func main() {
	flag.Parse()
	r := *RAND * rand.Float64()
	base := r + *BASE
	log.Printf("base: %.1f", base)

	tg := ToneGen{
		SampleRate: *RATE,
		ToneLen:    *SECS,
		RampLen:    *RAMP,
		BaseHz:     base,
		StepHz:     *STEP,
	}

	w := bufio.NewWriter(os.Stdout)
	switch *MODE {
	case "nested":
		fmt.Fprintf(os.Stderr, "%v\n", ExpandNested(W6REK))
		fmt.Fprintf(os.Stderr, "%v\n", ExpandWord(W6REK))
		fmt.Fprintf(os.Stderr, "%v\n", len(ExpandNested(W6REK)))
		fmt.Fprintf(os.Stderr, "%v\n", len(ExpandWord(W6REK)))

		volts := tg.Play(ExpandNested(W6REK))
		w.Write(VoltsToS16be(volts, *GAIN))

	case "chevron":
		down := tg.Boop(2, -1)
		w.Write(VoltsToS16be(down, *GAIN))
		w.Write(VoltsToS16be(down, *GAIN))
		w.Write(VoltsToS16be(down, *GAIN))

		word := tg.Play(ExpandWord(W6REK))
		w.Write(VoltsToS16be(word, *GAIN))

		up := tg.Boop(-1, 2)
		w.Write(VoltsToS16be(up, *GAIN))
		w.Write(VoltsToS16be(up, *GAIN))
		w.Write(VoltsToS16be(up, *GAIN))

	default:
		panic(*MODE)
	}
	w.Flush()
}
