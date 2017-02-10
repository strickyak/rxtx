/* Routines for receiving & transmitting & processing audio.
 */
package rxtx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"
)
import "github.com/strickyak/rxtx/mulaw"

const (
	PowerFlag = 0x01
	TxFlag    = 0x02
	AudioFlag = 0x04
	WhoFlag   = 0x08
)

const VERSION = 11411
const SamplesPerSec = 8000
const SamplesPerPacket = 200
const PacketsPerSecond = SamplesPerSec / SamplesPerPacket

var ENDIAN = binary.BigEndian
var HSIZE = binary.Size(new(Header))

func init() {
	if HSIZE < 0 {
		panic("HSIZE")
	}
}

// Fixed size SegmentQueue.  If you add too many,
// it drops the oldest ones.
const QLEN = PacketsPerSecond
type SegmentQueue struct {
	Vec [][]byte
	Begin	int
	End	int
	Size	int
}

func NewSegmentQueue() *SegmentQueue {
	return &SegmentQueue{Vec: make([][]byte, QLEN)}
}

func (q SegmentQueue) String() string {
	return fmt.Sprintf("%#v", q)
}
func (q *SegmentQueue) Add(segment []byte) {
	q.Begin = (q.Begin + 1) % QLEN
	q.Vec[q.Begin] = segment
	if q.Size == QLEN {
		// Drops the segment at the End.
		q.End = (q.End + 1) % QLEN
	} else {
		q.Size += 1
	}
}

func (q *SegmentQueue) Take() []byte {
	if q.Size == 0 {
		panic(q.String())
	} else {
		q.Size--
		q.End = (q.End + 1) % QLEN
		return q.Vec[q.End]
	}
}

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
	Touch time.Time
	Addr   *net.UDPAddr
	Skew	time.Duration  // Station's time minus our time, maximized.
}

type Socket struct {
	Addr   *net.UDPAddr
	Conn   *net.UDPConn
	Engine *Engine
}

type Engine struct {
	Me       int
	Other    int
	Stations map[string]*Station
	Sock      *Socket
	ProxyAddr *net.UDPAddr
	Audio     *os.File
	// Generating time & seq numbers.
	PrevTime	int64
	PrevSeq		int
}

func NewEngine(me int, proxyAddrString string) *Engine {
	a, err := net.ResolveUDPAddr("udp", proxyAddrString)
	if err != nil {
		panic(err)
	}
	return &Engine{
		Me:        me,
		Stations:  make(map[string]*Station),
		ProxyAddr: a,
		PrevTime: time.Now().Unix(),
	}
}

func (e *Engine) GetTimeAndSeq() (int64, int) {
	now := time.Now().Unix()
	if now == e.PrevTime {
		e.PrevSeq++
	} else {
		e.PrevSeq = 0
		e.PrevTime = now
	}
	return e.PrevTime, e.PrevSeq
}

func (e *Engine) FindStation(addr *net.UDPAddr) *Station {
	whom := addr.String()
	station, ok := e.Stations[whom]
	if !ok {
		station = &Station{
			Addr: addr,
		}
		e.Stations[whom] = station
	}
	station.Touch = time.Now()
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
	h := &Header{
		Version: VERSION,
		Length:  0,
		Flags:   0, // TODO
		Source:  byte(254),
		Dest:    byte(255),
	}
	t, s := e.GetTimeAndSeq()
	h.Time, h.Seq = int32(t), byte(s)
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
	println(ShowBytes(z[:sz]))

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

	if size < HSIZE {
		panic(size)
	}

	r := bytes.NewReader(packet)
	h := new(Header)
	err = binary.Read(r, binary.BigEndian, h)
	if err != nil {
		panic(err)
	}

	actualPayload := int(size) - HSIZE

	if actualPayload != int(h.Length) {
		log.Panicf("Got %d payload bytes; expected %d", actualPayload, h.Length)
	}

	if h.Length > 0 {
		if h.Length != SamplesPerPacket {
			log.Panicf("Got payload length %d wanted %d", h.Length, SamplesPerPacket)
		}
		copy(segment, packet[HSIZE:])
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
	for {
		size, addr, err := e.Sock.Conn.ReadFromUDP(packet)
		if err != nil {
			panic(err)
		}
		// println("PROXY GOT", size, addr, ShowBytes(packet[:20]))
		st0 := e.FindStation(addr)

		for _, st := range e.Stations {
			if st != st0 {
				out := packet[:size]
				_, err := e.Sock.Conn.WriteToUDP(out, st.Addr)
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

func (e *Engine) RadioCommand(usb string) {
	go e.RunRadioReceiveAudio(usb)
	e.RunRadioSendAudio()
}

func (e *Engine) RunRadioReceiveAudio(usb string) {
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
			n, err = dev.Write([]byte(usb))
			if err != nil {
				panic(err)
			}
			if n != len(usb) {
				log.Panicf("dev.Write wrote %d bytes, wanted %d", n, len(usb))
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
			SayPower(segment)
		} else {
			print(".")
		}
	}
}

func SayPower(segment []byte) {
	var sumsq int64
	var prev int64
	for _, e := range segment {
		a := int64(mulaw.DecodeMulaw16(e))
		diff := a - prev
		sumsq += diff * diff
		prev = a
	}
	x := math.Sqrt(float64(sumsq) / float64(len(segment)))
	fmt.Fprintf(os.Stderr, "(%d) ", int64(x))
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
		SayPower(segment)
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
			log.Fatalf("os.Stdin.Read < 1")
			os.Exit(0)
		}
		if err != nil {
			log.Fatalf("os.Stdin.Read --> err")
			os.Exit(2)
		}
		o.LastEnter = time.Now()
	}
}
