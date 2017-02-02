// +build main

package main

import (
	"flag"
	"strconv"
	"strings"

	. "github.com/strickyak/rxtx"
)

var bind = flag.String("bind", "localhost:1500", "port to bind to")
var peer = flag.String("peer", "2=localhost:1501", "port to connect to")
var me = flag.Int("me", 1, "my ID")
var cmd = flag.String("cmd", "t", "what to do")
var audio = flag.String("audio", "/dev/audio", "muLaw audio device")

func main() {
  e := NewEngine(*me)
  e.InitSocket(*bind)
  e.InitAudio(*audio)
  vec := strings.Split(*peer, "=")
  peerId, err := strconv.ParseInt(vec[0], 10, 64)
  if err != nil { panic(err) }
  e.RegisterStation(byte(peerId), vec[1])
  switch *cmd {
  case "t": e.Transmit()
  //case "r": e.Receive()
  default: panic(*cmd)
  }
}
