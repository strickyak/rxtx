// +build main

/*
   Beeps every other second, to demonstrate you can just not
   write to /dev/audio for a while to create gaps.

   Usage:  go run beep.go > /dev/audio

   Hint:
       sudo modprobe snd_pcm_oss
       sudo modprobe snd_mixer_oss
*/
package main

import (
	"os"
	"time"
)

var Buf = make([]byte, 8000)

func init() {
	for i := 0; i < 8000; i++ {
		if (i & 16) == 0 {
			Buf[i] = 0xC0
		} else {
			Buf[i] = 0x40
		}
	}
}
func Beep() {
	n, err := os.Stdout.Write(Buf)
	if err != nil {
		println(err)
		panic(err)
	}
	if n != len(Buf) {
		println(n)
		panic(n)
	}
}

func main() {
	for {
		os.Stderr.Write([]byte("on\n"))
		go Beep()
		time.Sleep(time.Second)
		os.Stderr.Write([]byte("off\n"))
		time.Sleep(time.Second)
	}
}
