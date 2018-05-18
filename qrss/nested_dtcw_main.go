// +build main

// go run nested_dtcw_main.go | pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100 --device=alsa_output.usb-Burr-Brown_from_TI_USB_Audio_CODEC-00.analog-stereo
package main

import (
	"bufio"
	"fmt"
	"os"
)
import "github.com/strickyak/rxtx/qrss"

func main() {
	fmt.Fprintf(os.Stderr, "%v\n", qrss.ExpandNested(qrss.W6REK))
	fmt.Fprintf(os.Stderr, "%v\n", qrss.ExpandWord(qrss.W6REK))

	bb := qrss.Play(qrss.ExpandNested(qrss.W6REK))
	w := bufio.NewWriter(os.Stdout)
	w.Write(bb)
	w.Flush()
}
