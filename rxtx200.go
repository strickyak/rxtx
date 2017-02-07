// +build main

// sudo ufw allow 44144/udp
// ./rxtx-200 --bind=:44144 --cmd=p
// go run main.go --proxy=xxxxx.yyy.zed:44144

package main

import (
	"flag"
	//"strconv"
	//"strings"

	. "github.com/strickyak/rxtx"
)

var bind = flag.String("bind", ":44144", "port to bind to")
var proxy = flag.String("proxy", "forth.yak.net:1500", "proxy to use")
var me = flag.Int("me", 1, "my ID")
var cmd = flag.String("cmd", "h", "what to do")
var audio = flag.String("audio", "/dev/audio", "muLaw audio device")
var junk = flag.String("junk", "abcdefgh", "junk to write on usb")

func main() {
	flag.Parse()
	e := NewEngine(*me, *proxy)
	e.InitSocket(*bind)
	switch *cmd {

	case "p": // proxy
		e.ProxyCommand()

	case "h": // human
		e.InitAudio(*audio)
		e.HumanCommand()

	case "r": // radio
		e.InitAudio(*audio)
		e.RadioCommand(*junk)

	default:
		panic(*cmd)
	}
}
