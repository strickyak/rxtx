/*
   Routines for receiving & transmitting & processing audio.
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
	Seq    byte
	Flags  byte
	Source byte
	Dest   byte
	// TODO -- authenticate after payload.
}

type Station struct {
	Id byte
	// Time time.Time   // Latest time from this station.
	// Seq  byte        // Latest sequence from this station.
	Addr *net.UDPAddr
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
	Sock *Socket
	Peer *Station
	Audio *Audio
}

func NewEngine(me int) *Engine {
  return &Engine {
    Me: me,
    Stations: make(map[byte]*Station),
  }
}

func (e *Engine) RegisterStation(id byte, addr string) *Station {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		panic(err)
	}
	station := &Station{
		Id: id,
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

func (e *Engine) WritePacket(h *Header, b []byte) {
	// now := time.Now().Unix()
	sz := binary.Size(h)
	z := make([]byte, sz + len(b))
	w := bytes.NewBuffer(z)
	err := binary.Write(w, ENDIAN, h)
	if err != nil {
		panic(err)
	}
	copy(z[sz:], b)
	station, ok := e.Stations[h.Dest]
	if !ok {
	  log.Panicf("WritePacket: station not registered: %d", h.Dest)
	}
	n, err := e.Sock.Conn.WriteToUDP(z, station.Addr)
	if err != nil {
	  log.Panicf("WritePacket: Cannot WriteToUDP: %v", err)
	}
	if n != len(z) {
	  log.Panicf("WritePacket: Wrote %d bytes, expected %d", n, len(z))
	}
}

func (e *Engine) ReadPacket(b []byte) (*Header, *Station) {
	b = make([]byte, 1024)
	size, addr, err := e.Sock.Conn.ReadFromUDP(b)
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

	r := bytes.NewReader(b)
	err = binary.Read(r, binary.BigEndian, h)
	if err != nil {
		panic(err)
	}

	// TODO: Authenticate.
	actualPayload := int(size) - hSize
	if int(size)-hSize != int(h.Length) {
		Panicf("Got %d payload bytes; expected %d", actualPayload, h.Length)
	}

	station, ok := e.Stations[h.Source]
	if !ok {
		station := &Station{
			Id: h.Source,
			Addr: addr,
		}
		e.Stations[h.Source] = station
	}

	return h, station
}

type Audio struct {
  File *os.File
}

func (e *Engine) InitAudio(path string) {
  fd, err := os.OpenFile(path, os.O_RDWR, 0666)
  if err != nil { panic(err) }
  e.Audio = &Audio{
    File: fd,
  }
}

func (e *Engine) Transmit() {
}
