/* Routines for receiving & transmitting & processing audio.
*/
package rxtx

import (
	"bytes"
	"encoding/binary"
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
const PacketsPerSec = 20
const SamplesPerPacket = 400

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

func (e *Engine) MarkStation(h *Header, addr *net.UDPAddr) {
	st, ok := e.Stations[h.Source]
	if !ok {
		st = &Station{
			Id: h.Source,
		}
		e.Stations[h.Source] = st
	}
	st.Latest = time.Now()
	st.Seq = 0
	st.Addr = addr
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

func (e *Engine) WritePacket(h *Header, b []byte, dest *net.UDPAddr) {
	h.Length = SamplesPerPacket
	sz := binary.Size(h)
	z := make([]byte, sz+len(b))
	w := bytes.NewBuffer(z)
	err := binary.Write(w, ENDIAN, h)
	if err != nil {
		panic(err)
	}
	copy(z[sz:], b)

	n, err := e.Sock.Conn.WriteToUDP(z, dest)

	if err != nil {
		log.Panicf("WritePacket: Cannot WriteToUDP: %v", err)
	}
	if n != len(z) {
		log.Panicf("WritePacket: Wrote %d bytes, expected %d", n, len(z))
	}
}

func (e *Engine) ReadPacket(b []byte) *Header {
	packet := make([]byte, 1024)
	size, addr, err := e.Sock.Conn.ReadFromUDP(packet)
	if err != nil {
		panic(err)
	}

	h := new(Header)
	hSize := binary.Size(h)
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
	e.MarkStation(h, addr)

	// TODO: Authenticate.
	actualPayload := int(size) - hSize
	if int(size)-hSize != int(h.Length) {
		Panicf("Got %d payload bytes; expected %d", actualPayload, h.Length)
	}

	if h.Length > 0 {
		if h.Length != SamplesPerPacket {
			Panicf("Got payload length %d wanted %d", h.Length, SamplesPerPacket)
		}
	}

	return h
}

func (e *Engine) InitAudio(path string) {
	audio, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	e.Audio = audio
}

func (e *Engine) Transmit() {
	log.Fatalf("obsolete")
}

func (e *Engine) Human() {
	ptt := new(PushToTalk)
	go ptt.Run()
	e.RunSendAudio(ptt)
	/*
		for {
		  if ptt.Active() {
		    print("Y")
		  } else {
		    print(".")
		  }
		  time.Sleep(time.Second / 10)
		}
	*/
}

func (e *Engine) SendToProxy(buf []byte) {
	h := e.ForgeHeader()
	e.WritePacket(h, buf, e.ProxyAddr)
}

func (e *Engine) RunSendAudio(ptt *PushToTalk) {
	buf := make([]byte, SamplesPerPacket)
	for {
		n, err := e.Audio.Read(buf)
		if err != nil {
			panic(err)
		}
		if n != SamplesPerPacket {
			log.Panicf("e.Audio.Read got %d bytes, wanted %d", n, SamplesPerPacket)
			os.Exit(13)
		}
		if ptt.Active() {
			e.SendToProxy(buf)
			print(":")
		} else {
			print(".")
		}
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
