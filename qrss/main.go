// +build main

// go run main.go | pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100 --device=alsa_output.usb-Burr-Brown_from_TI_USB_Audio_CODEC-00.analog-stereo
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
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

var MODE = flag.String("mode", "chevron", "nested | chevron")

func main() {
	flag.Parse()
	var r int
	if *RAND > 0 {
		r = Random(int(*RAND))
	}
	base := float64(r) + *BASE
	log.Printf("base: %f", base)

	tg := ToneGen{
		SampleRate: *RATE,
		ToneLen:    *SECS,
		RampLen:    *RAMP,
		BaseHz:     base,
		StepHz:     *STEP,
	}

	w := bufio.NewWriter(os.Stdout)
	done := make(chan bool)
	vv := make(chan Volt, 42)
	go EmitVolts(vv, *GAIN, w, done) // Consume channel vv of volts, writing stdout.

	// Produce on channel vv of volts.
	switch *MODE {
	case "nested": // Experimental.
		fmt.Fprintf(os.Stderr, "%v\n", ExpandNested(W6REK))
		fmt.Fprintf(os.Stderr, "%v\n", ExpandWord(W6REK))
		fmt.Fprintf(os.Stderr, "%v\n", len(ExpandNested(W6REK)))
		fmt.Fprintf(os.Stderr, "%v\n", len(ExpandWord(W6REK)))

		tg.Play(ExpandNested(W6REK), vv)

	case "chevron": // Standard.
		tg.Boop(2, -1, vv) // Descending tone, from level 2 to level -1.
		tg.Play(ExpandWord(W6REK), vv)
		tg.Boop(-1, 2, vv) // Ascending tone, from level -1 to level 2.

	default:
		panic(*MODE)
	}

	close(vv)
	<-done // Wait for EmitVolts to finish.
	w.Flush()
}
