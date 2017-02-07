/* Routines for receiving & transmitting & processing audio.
 */
package rxtx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)
import . "log"

const (
	PowerFlag = 0x01
	TxFlag    = 0x02
	AudioFlag = 0x04
	WhoFlag   = 0x08
)

const VERSION = 11411
const SamplesPerSec = 8000
const SamplesPerPacket = 200

var ENDIAN = binary.BigEndian

type Header struct {
	Version int16
	Length  int16
	Time    int32
	Seq     byte
	Flags   byte
	Source  byte
	Dest    byte
	// TODO -- authenticate after payload.
}

type Station struct {
	Id     byte
	Latest time.Time // Latest time from this station.
	Seq    byte      // Latest sequence from this station.
	Addr   *net.UDPAddr
}

type Socket struct {
	Addr   *net.UDPAddr
	Conn   *net.UDPConn
	Engine *Engine
}

type Engine struct {
	Me       int
	Other    int
	Stations map[byte]*Station
	// PrevSecs int
	// PrevSeq  int
	Sock      *Socket
	ProxyAddr *net.UDPAddr
	Audio     *os.File
}

func NewEngine(me int, proxyAddrString string) *Engine {
	a, err := net.ResolveUDPAddr("udp", proxyAddrString)
	if err != nil {
		panic(err)
	}
	return &Engine{
		Me:        me,
		Stations:  make(map[byte]*Station),
		ProxyAddr: a,
	}
}

func (e *Engine) RegisterStation(id byte, addr string) *Station {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		panic(err)
	}
	station := &Station{
		Id:   id,
		Addr: a,
	}
	e.Stations[id] = station
	return station
}

func (e *Engine) InitSocket(localAddr string) {
	addr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	e.Sock = &Socket{
		Addr:   addr,
		Conn:   conn,
		Engine: e,
	}
}

func (e *Engine) ForgeHeader() *Header {
	now := time.Now().Unix()
	h := &Header{
		Version: VERSION,
		Length:  0,
		Time:    int32(now),
		Seq:     0, // TODO
		Flags:   0, // TODO
		Source:  byte(e.Me),
		Dest:    byte(255),
	}
	return h
}

func (e *Engine) WritePacket(h *Header, segment []byte, dest *net.UDPAddr) {
	h.Length = int16(len(segment))
	sz := binary.Size(h)
	w := bytes.NewBuffer(nil)
	err := binary.Write(w, ENDIAN, h)
	if err != nil {
		panic(err)
	}
	z := make([]byte, sz+len(segment))
	copy(z[:sz], w.String())
	copy(z[sz:], segment)

	n, err := e.Sock.Conn.WriteToUDP(z, dest)

	if err != nil {
		log.Panicf("WritePacket: Cannot WriteToUDP: %v", err)
	}
	if n != len(z) {
		log.Panicf("WritePacket: Wrote %d bytes, expected %d", n, len(z))
	}
}

func (e *Engine) ReadPacket(segment []byte) *Header {
	packet := make([]byte, 512)
	size, _, err := e.Sock.Conn.ReadFromUDP(packet)
	if err != nil {
		panic(err)
	}
	println("size", size)

	h := new(Header)
	hSize := binary.Size(h)
	println("hSize", hSize)
	if hSize < 0 {
		panic("hSize")
	}

	if size < hSize {
		panic(size)
	}

	r := bytes.NewReader(packet)
	err = binary.Read(r, binary.BigEndian, h)
	if err != nil {
		panic(err)
	}
	// e.MarkStation(h, addr)

	// TODO: Authenticate.
	actualPayload := int(size) - hSize
	println("actualPayload", actualPayload)
	fmt.Printf("%#v\n", *h)
	/*
		if actualPayload != int(h.Length) {
			Panicf("Got %d payload bytes; expected %d", actualPayload, h.Length)
		}

		if h.Length > 0 {
			if h.Length != SamplesPerPacket {
				Panicf("Got payload length %d wanted %d", h.Length, SamplesPerPacket)
			}
		}
	*/

	copy(segment, packet[hSize:])

	return h
}

func (e *Engine) InitAudio(path string) {
	audio, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	e.Audio = audio
}

func ShowBytes(bb []byte) string {
	z := bytes.NewBufferString("[")
	prev := -1
	for _, e := range bb {
		if int(e) == prev {
			z.WriteRune('*')
		} else {
			fmt.Fprintf(z, "%d,", e)
		}
		prev = int(e)
	}
	z.WriteRune(']')
	return z.String()
}

func (e *Engine) ProxyCommand() {
	packet := make([]byte, 512)
	pals := make(map[string]*net.UDPAddr)
	for {
		size, addr, err := e.Sock.Conn.ReadFromUDP(packet)
		if err != nil {
			panic(err)
		}
		// println("PROXY GOT", size, addr, ShowBytes(packet[:20]))

		whom := addr.String()
		pals[whom] = addr

		for w, a := range pals {
			if w != whom || w == whom {
				out := packet[:size]
				_, err := e.Sock.Conn.WriteToUDP(out, a)
				// println("PROXY WROTE", n, a, ShowBytes(out))
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (e *Engine) HumanCommand() {
	ptt := new(PushToTalk)
	go ptt.Run()
	go e.RunReceiveAudio()
	e.RunSendAudio(ptt)
}

func (e *Engine) RadioCommand() {
	go e.RunRadioReceiveAudio()
	e.RunRadioSendAudio()
}

func (e *Engine) RunRadioReceiveAudio() {
	dev, devErr := os.OpenFile("/dev/ttyUSB0", os.O_RDWR, 0666)
	if devErr != nil {
		panic(devErr)
	}

	segment := make([]byte, SamplesPerPacket)
	for {
		h := e.ReadPacket(segment)
		if h.Length > 0 {
			n, err := e.Audio.Write(segment)
			if err != nil {
				panic(err)
			}
			if n != SamplesPerPacket {
				log.Panicf("e.Audio.Write wrote %d bytes, wanted %d", n, SamplesPerPacket)
				os.Exit(13)
			}

			// Send junk on Serial Cable to cause PTT on Radio.
			junk := "abc"
			n, err = dev.Write([]byte(junk))
			if err != nil {
				panic(err)
			}
			if n != len(junk) {
				log.Panicf("dev.Write wrote %d bytes, wanted %d", n, len(junk))
				os.Exit(13)
			}
		}
	}
}

func (e *Engine) RunReceiveAudio() {
	segment := make([]byte, SamplesPerPacket)
	for {
		h := e.ReadPacket(segment)
		if h.Length > 0 {
			n, err := e.Audio.Write(segment)
			if err != nil {
				panic(err)
			}
			if n != SamplesPerPacket {
				log.Panicf("e.Audio.Write wrote %d bytes, wanted %d", n, SamplesPerPacket)
				os.Exit(13)
			}
		}
	}
}

func (e *Engine) SendToProxy(segment []byte) {
	h := e.ForgeHeader()
	e.WritePacket(h, segment, e.ProxyAddr)
}

func (e *Engine) RunSendAudio(ptt *PushToTalk) {
	segment := make([]byte, SamplesPerPacket)
	for {
		n, err := e.Audio.Read(segment)
		if err != nil {
			panic(err)
		}
		if n != SamplesPerPacket {
			log.Panicf("e.Audio.Read got %d bytes, wanted %d", n, SamplesPerPacket)
			os.Exit(13)
		}
		if ptt.Active() {
			e.SendToProxy(segment)
			print(":")
		} else {
			print(".")
		}
	}
}

func (e *Engine) RunRadioSendAudio() {
	segment := make([]byte, SamplesPerPacket)
	for {
		n, err := e.Audio.Read(segment)
		if err != nil {
			panic(err)
		}
		if n != SamplesPerPacket {
			log.Panicf("e.Audio.Read got %d bytes, wanted %d", n, SamplesPerPacket)
			os.Exit(13)
		}
		e.SendToProxy(segment)
	}
}

type PushToTalk struct {
	LastEnter time.Time
}

var enterToTalkBuf = make([]byte, 2)

func (o *PushToTalk) Active() bool {
	return o.LastEnter.Add(time.Second).After(time.Now())
}
func (o *PushToTalk) Run() {
	for {
		n, err := os.Stdin.Read(enterToTalkBuf)
		if n < 1 {
			println("os.Stdin.Read < 1")
			os.Exit(0)
		}
		if err != nil {
			println("os.Stdin.Read --> err")
			os.Exit(2)
		}
		o.LastEnter = time.Now()
	}
}
