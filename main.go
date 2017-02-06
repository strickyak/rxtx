// +build main

package main

import (
	"flag"
	//"strconv"
	//"strings"

	. "github.com/strickyak/rxtx"
)

var bind = flag.String("bind", "localhost:1500", "port to bind to")
var proxy = flag.String("proxy", "forth.yak.net:1500", "proxy to use")
var me = flag.Int("me", 1, "my ID")
var cmd = flag.String("cmd", "h", "what to do")
var audio = flag.String("audio", "/dev/audio", "muLaw audio device")

func main() {
	flag.Parse()
	e := NewEngine(*me, *proxy)
	e.InitSocket(*bind)
	switch *cmd {
	case "T":
		e.InitAudio(*audio)
		e.Transmit()
	//case "T": e.Receive()
	//case "p": e.Proxy()
	case "h":
		e.InitAudio(*audio)
		e.Human()
	//case "r": e.Radio()
	default:
		panic(*cmd)
	}
}
