/* PushToTalk app, server (proxy), & radio controller.

Next to do:
  Don't send radio audio back to radio.

*/
package rxtx

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/strickyak/rxtx/mulaw"
)

var ORIGIN = flag.Int("o", 1, "bitset")
var CHANNEL = flag.Int("chan", 1, "bitset")
var SUBSCRIBE = flag.Int("sub", 1, "bitset")
var SQUELCH = flag.Int("squelch", 300, "mulaw mean-square")

/*
	Packets per second, for 8000B/sec:
	  50x 160B = 20ms
	  40x 200B = 25ms
	  32x 250B = 31ms
	  25x 320B = 40ms
	  20x 400B = 50ms
*/

const VERSION = 11411
const SamplesPerSecond = 8000
const N = 200 // N == Samples per Second
const PacketsPerSecond = SamplesPerSecond / N
const BytesPerSampleInPayload = 1
const PayloadLen = N * BytesPerSampleInPayload

var ENDIAN = binary.BigEndian
var HSIZE = binary.Size(new(Header))

func init() {
	// Check that Header is compatable with binary encoding.
	if HSIZE < 0 {
		panic("HSIZE")
	}
}

type FlagBits byte

func (b FlagBits) Has(f FlagBits) bool {
	return (b & f) == f
}

const (
	AudioFlag FlagBits = (1 << iota) // Contains Audio Payload.
	KeepAliveFlag
	RadioFlag
	FullDuplexFlag
)

type Header struct {
	Version           int16
	Flags             FlagBits
	Origin            BitSet // User(s) who are speaking.
	Channel           BitSet // Packet goes to these channels.
	Subscribe         BitSet // User subscribes to these channels.
	UnixNanoTimestamp int64
}

func (h Header) String() string {
	return fmt.Sprintf("{V%d F%d O:%d C:%d S:%d TS:%d} ",
		h.Version, h.Flags, h.Origin, h.Channel, h.Subscribe, h.UnixNanoTimestamp)
}

type Packet struct {
	Header  *Header
	Segment []byte
	Ints    []int16
}

type Station struct {
	Id    string
	Touch time.Time
	Addr  *net.UDPAddr
	Skew  time.Duration // Station's time minus our time, maximized.
	Queue *PacketQueue  // Received from the station.

	Subscribe BitSet // User subscribes to these channels.
	Origin    BitSet // User(s) who are speaking.
	Channel   BitSet // Packet goes to these channels.
}

type Socket struct {
	Addr   *net.UDPAddr
	Conn   *net.UDPConn
	Engine *Engine
}

type Engine struct {
	Stations map[string]*Station
	Sock     *Socket
	HubAddr  *net.UDPAddr
	Audio    *os.File
}

func NewEngine(proxyAddrString string) *Engine {
	a, err := net.ResolveUDPAddr("udp", proxyAddrString)
	if err != nil {
		panic(err)
	}
	return &Engine{
		Stations: make(map[string]*Station),
		HubAddr:  a,
	}
}

func (e *Engine) FindStation(addr *net.UDPAddr, h *Header) *Station {
	whom := addr.String()
	station, ok := e.Stations[whom]
	if !ok {
		station = &Station{
			Id:    addr.String(),
			Addr:  addr,
			Queue: NewPacketQueue(),
			Skew:  -356 * 86400 * time.Second,
		}
		e.Stations[whom] = station
	}
	station.Touch = time.Now()
	if h != nil {
		station.Origin = h.Origin
		station.Subscribe = h.Subscribe
		station.Channel = h.Channel
	}
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
		Version:   VERSION,
		Flags:     AudioFlag,
		Origin:    BitSet(*ORIGIN),
		Channel:   BitSet(*CHANNEL),
		Subscribe: BitSet(*SUBSCRIBE),
	}
	h.UnixNanoTimestamp = time.Now().UnixNano()
	return h
}

func (e *Engine) WritePacket(h *Header, segment []byte, dest *net.UDPAddr) {
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

func (e *Engine) ReadPacket(segment []byte, conn *net.UDPConn) (*Header, *net.UDPAddr) {
	packet := make([]byte, 512)
	size, addr, err := conn.ReadFromUDP(packet)
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

	if (h.Flags & AudioFlag) == AudioFlag {
		if actualPayload != N {
			log.Printf("Got payload length %d wanted %d", PayloadLen, N)
		}
		copy(segment, packet[HSIZE:])
	}
	return h, addr
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

func (e *Engine) SendToEach(p *Packet) {
	println("PROXY SendToEach", ShowBytes(p.Segment[:20]))
	for addr, st := range e.Stations {
		println("Station:", addr, time.Since(st.Touch).String())
		if time.Since(st.Touch) > 30*time.Second {
			println("BAD TIME", st.Touch.String())
			continue
		}

		if st.Subscribe&p.Header.Channel == 0 {
			continue
		}

		h := e.ForgeHeader()
		h.Origin = p.Header.Origin
		h.Channel = p.Header.Channel
		h.Subscribe = p.Header.Subscribe
		println("WRITE TO Station:", addr)
		e.WritePacket(h, p.Segment, st.Addr)
	}
}

func MulawDecode(a []byte) []int16 {
	z := make([]int16, N)
	for i := 0; i < N; i++ {
		z[i] += mulaw.DecodeMulaw16(a[i])
	}
	return z
}
func MulawEncode(a []int16) []byte {
	z := make([]byte, N)
	for i := 0; i < N; i++ {
		z[i] = mulaw.EncodeMulaw16(a[i])
	}
	return z
}

func (e *Engine) Sendem() {
	var origins BitSet
	packets := make(map[string]*Packet)

	// Collect a segment from very station.
	for _, st := range e.Stations {
		for packet := st.Queue.Take(); packet != nil; packet = st.Queue.Take() {
			if packet.Header.Flags&AudioFlag == 0 {
				log.Fatalln("non-audio packets should already be handled", packet.Header.String())
				continue
			}
			packets[st.Id] = packet
			packet.Ints = MulawDecode(packet.Segment)
			origins |= st.Origin
			break // Only take 1 audio packet off the Queue at a time.
		}
	}

	// Gather and send to every station.
	for id, st := range e.Stations {
		sum := make([]int16, N)
		something := false

		// Examine & include all packets going to the station.
		for id, p := range packets {
			if (id == st.Id) && !(p.Header.Flags.Has(FullDuplexFlag)) {
				continue
			}

			// Adds the segment to the sum, setting something.
			for i := 0; i < N; i++ {
				x := p.Ints[i]
				sum[i] += x
				if x != 0 {
					something = true
				}
			}
		}

		if something {
			h := e.ForgeHeader()
			h.Origin = origins
			h.Channel = BitSet(0xFF)
			h.Subscribe = BitSet(0xFF)
			println("WRITE TO Station:", id)
			e.WritePacket(h, MulawEncode(sum), st.Addr)
		}
	}
}

func (e *Engine) HubSendLoop() {
	t0 := time.Now()
	delta := time.Second / PacketsPerSecond
	i := 0
	for {
		time.Sleep(time.Millisecond)
		now := time.Now()
		target := t0.Add(time.Duration(i) * delta)
		if now.Before(target) {
			continue
		}
		e.Sendem()
	}
}

func (e *Engine) HubRecvLoop() {
	for {
		segment := make([]byte, N)
		h, addr := e.ReadPacket(segment, e.Sock.Conn)
		st := e.FindStation(addr, h)

		packetTime := time.Unix(0, h.UnixNanoTimestamp)
		skew := packetTime.Sub(time.Now())
		if skew < -5*time.Second {
			continue // Don't tolerate much skew.
		}

		if st.Skew < skew {
			st.Skew = skew
		}

		if h.Flags == KeepAliveFlag {
			fmt.Fprintf(os.Stderr, " <KEEPALIVE:%d:%v> ", h.Flags, addr)
		}

		st.Queue.Add(&Packet{h, segment, nil})
	}
}

func (e *Engine) HubCommand() {
	go e.HubRecvLoop()
	e.HubSendLoop()
}

func (e *Engine) HumanCommand() {
	ptt := new(PushToTalk)
	go ptt.Run()
	go e.KeepAliveLoop(KeepAliveFlag)
	go e.HumanReceiveNetworkLoop(ptt)
	e.HumanSendNetworkLoop(ptt)
}

func (e *Engine) RadioCommand(usb string) {
	go e.KeepAliveLoop(RadioFlag | KeepAliveFlag)
	go e.RadioReceiveNetworkLoop(usb)
	e.RadioSendNetworkLoop()
}

func (e *Engine) RadioSendNetworkLoop() {
	segment := make([]byte, N)
	for {
		n, err := e.Audio.Read(segment)
		Check(err)
		if n != N {
			log.Panicf("e.Audio.Read got %d bytes, wanted %d", n, N)
			os.Exit(13)
		}
		x := SayPower(segment)
		if x > float64(*SQUELCH) {
			e.SendToHub(segment, RadioFlag)
		}
	}
}

func (e *Engine) RadioReceiveNetworkLoop(usb string) {
	var lastTimestamp int64

	dev, devErr := os.OpenFile("/dev/ttyUSB0", os.O_RDWR, 0666)
	Check(devErr)

	segment := make([]byte, N)
	for {
		h, _ := e.ReadPacket(segment, e.Sock.Conn)
		_ = h
		if PayloadLen > 0 && h.UnixNanoTimestamp > lastTimestamp {
			lastTimestamp = h.UnixNanoTimestamp

			n, err := e.Audio.Write(segment)
			Check(err)
			Assert(n == N)

			// Send junk on Serial Cable to cause PTT on Radio.
			n, err = dev.Write([]byte(usb))
			Check(err)
			Assert(n == len(usb))
		}
	}
}

func (e *Engine) HumanReceiveNetworkLoop(ptt *PushToTalk) {
	segment := make([]byte, N)
	for {
		h, _ := e.ReadPacket(segment, e.Sock.Conn)
		_ = h
		if PayloadLen > 0 && !ptt.Active() {
			n, err := e.Audio.Write(segment)
			if err != nil {
				panic(err)
			}
			if n != N {
				log.Panicf("e.Audio.Write wrote %d bytes, wanted %d", n, N)
				os.Exit(13)
			}
		}
	}
}

func (e *Engine) SendToHub(segment []byte, extraFlags FlagBits) {
	h := e.ForgeHeader()
	h.Flags |= extraFlags
	e.WritePacket(h, segment, e.HubAddr)
}

func (e *Engine) KeepAliveLoop(flags FlagBits) {
	for {
		h := e.ForgeHeader()
		h.Flags = flags
		e.WritePacket(h, nil, e.HubAddr)
		time.Sleep(10 * time.Second)
	}
}

func (e *Engine) HumanSendNetworkLoop(ptt *PushToTalk) {
	segment := make([]byte, N)
	for {
		n, err := e.Audio.Read(segment)
		if err != nil {
			panic(err)
		}
		if n != N {
			log.Panicf("e.Audio.Read got %d bytes, wanted %d", n, N)
			os.Exit(13)
		}
		if ptt.Active() {
			_ = SayPower(segment)
			e.SendToHub(segment, 0)
		} else {
			print(".")
		}
	}
}

func SayPower(segment []byte) float64 {
	var sumsq int64
	var prev int64
	for _, e := range segment {
		a := int64(mulaw.DecodeMulaw16(e))
		diff := a - prev
		sumsq += diff * diff
		prev = a
	}
	// Actually the square of the power (it's MS, not RMS).
	x := float64(sumsq) / float64(len(segment))
	fmt.Fprintf(os.Stderr, "(%d) ", int64(x))
	return x
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
		Check(err)
		if n < 1 {
			log.Fatalf("os.Stdin.Read < 1")
			os.Exit(0)
		}
		o.LastEnter = time.Now()
	}
}
