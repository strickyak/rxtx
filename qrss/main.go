// +build main

// go run main.go | pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100 --device=alsa_output.usb-Burr-Brown_from_TI_USB_Audio_CODEC-00.analog-stereo
// go run main.go --mode=raster --tag=hope --rate=44100 --secs=2 --ramp=0.1 | pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100
// go run main.go --mode=slashup --tag=w2h --rate=44100  --secs=1 --ramp=0.1 --step=100 | pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100

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
var RAND = flag.Float64("base_rand", 0, "Random addition to Base Hz")
var STEP = flag.Float64("step", 4, "Tone Step Hz")

var MODE = flag.String("mode", "chevron", "nested | chevron | slanted")
var TAG = flag.String("tag", "w6rek", "w6rek | w6rek/4/atl")

func main() {
	flag.Parse()
	var r int
	if *RAND > 0 {
		r = Random(int(*RAND))
	}
	base := float64(r) + *BASE
	log.Printf("base: %f", base)

	var tag []string
	switch *TAG {
	case "w6rek":
		tag = W6REK
	case "w6rek/4":
		tag = W6REK_4
	case "w6rek/4/atl":
		tag = W6REK_4_ATL
	case "hope":
	case "w2h":
		tag = W2H
		*TAG = "hackers W*H on W*H planet W*H earth W*H hope.net W*H DE W6REK "
	case "hny":
		*TAG = "HappyNewYear EM91nx W6REK/4 Pax"
		*MODE = "raster"
	default:
		tag = W6REK
	}

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
	case "nested":
		fmt.Fprintf(os.Stderr, "%v\n", ExpandNested(tag))
		fmt.Fprintf(os.Stderr, "%v\n", ExpandWord(tag))
		fmt.Fprintf(os.Stderr, "%v\n", len(ExpandNested(tag)))
		fmt.Fprintf(os.Stderr, "%v\n", len(ExpandWord(tag)))

		tg.PlayTones(ExpandNested(tag), vv)

	case "chevron":
		tg.Boop(2, -1, Both, vv) // Descending tone, from level 2 to level -1.
		tg.PlayTones(ExpandWord(tag), vv)
		tg.Boop(-1, 2, Both, vv) // Ascending tone, from level -1 to level 2.

	case "slashup": // e.g. "W/2/H" e.g. "_--/__---/____"
		for i, didahs := range tag {
			if i > 0 {
				tg.Boop(0.5, 2.5, Both, vv) // Ascending tone
			}
			for _, d := range didahs {
				switch d {
				case '.':
					tg.Boop(1, 1, Both, vv)
				case '-':
					tg.Boop(2, 2, Both, vv)
				default:
					panic("WUT:" + didahs)
				}
			}
		}

	case "slanted":
		tg.PlayTonePairs(SlantedExpandWord(tag), vv)

	case "neo":
		tg.PlayTonePairs(NeoExpandWord(tag), vv)
	case "duo":
		tg.PlayTonePairs(DuoExpandWord(tag), vv)
	case "jots":
		PrintJots(J_W6REK)
		tg.PlayTonePairs(JotsExpandWord(J_W6REK), vv)
	case "raster":
		tg.PlayTonePairs(RasterExpandWord(*TAG), vv)
	default:
		panic(*MODE)
	}

	close(vv)
	<-done // Wait for EmitVolts to finish.
	w.Flush()
}
