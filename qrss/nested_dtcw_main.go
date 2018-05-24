// +build main

// go run nested_dtcw_main.go | pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100 --device=alsa_output.usb-Burr-Brown_from_TI_USB_Audio_CODEC-00.analog-stereo
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)
import . "github.com/strickyak/rxtx/qrss"

var RATE = flag.Float64("rate", 44100, "Audio Sample Rate")
var SECS = flag.Float64("secs", 6, "Tone length in secs")
var RAMP = flag.Float64("ramp", 1.0, "Tone length in secs")
var GAIN = flag.Float64("gain", 0.86, "Modulation")

func main() {
	tg := ToneGen{
		SampleRate: *RATE,
		ToneLen:    *SECS,
		RampLen:    *RAMP,
	}

	fmt.Fprintf(os.Stderr, "%v\n", ExpandNested(W6REK))
	fmt.Fprintf(os.Stderr, "%v\n", ExpandWord(W6REK))
	fmt.Fprintf(os.Stderr, "%v\n", len(ExpandNested(W6REK)))
	fmt.Fprintf(os.Stderr, "%v\n", len(ExpandWord(W6REK)))

	volts := tg.Play(ExpandNested(W6REK))
	w := bufio.NewWriter(os.Stdout)
	w.Write(VoltsToS16be(volts, *GAIN))
	w.Flush()
}
