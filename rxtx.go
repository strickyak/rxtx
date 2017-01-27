/*
   Routines for receiving & transmitting & processing audio.
*/
package rxtx

//import (
//	"encoding/binary"
//	"net"
//)
//import . "log"
//
//const (
//	PowerFlag = 0x01
//	TxFlag    = 0x02
//	AudioFlag = 0x04
//	WhoFlag   = 0x08
//
//	OperatorNode = 1
//	ProxyNode    = 101
//	RadioNode    = 201
//)
//
//type Header struct {
//	Version int16
//	Length  int16
//	Time    int32
//
//	Check int64
//
//	Seq    int8
//	Flags  int8
//	Source int8
//	Dest   int8
//}
//
//const HeaderLen = 20
//
//type Host struct {
//	Node int8
//	Addr *net.UDPAddr
//}
//
//var Hosts = make(map[int8]Host)
//
//type Socket struct {
//	Addr *net.UDPAddr
//	Conn *net.UDPConn
//}
//
//func NewSocket(localAddr string) *Socket {
//	addr, err := net.ResolveUDPAddr("udp", localAddr)
//	if err != nil {
//		panic(err)
//	}
//	conn, err := net.ListenUDP("udp", addr)
//	if err != nil {
//		panic(err)
//	}
//	return &Socket{
//		Addr: addr,
//		Conn: conn,
//	}
//}
//
//func (sock *Socket) ReadFrom(b []byte) *Header {
//	b = new([]byte, 1024, 1024)
//	size, addr, err := sock.Conn.ReadFromUDP(b)
//	if err != nil {
//		panic(err)
//	}
//
//	h := new(Header)
//	hSize := binary.Size(h)
//	if hSize < 0 {
//		panic("hSize")
//	}
//
//	if size < hSize {
//		panic(size)
//	}
//
//	r := bytes.NewReader(b)
//	err = binary.Read(r, BigEndian, h)
//	if err != nil {
//		panic(err)
//	}
//
//	actualPayload := size - hSize
//	if size-hSize != h.Length {
//		Panicf("Got %d payload bytes; expected %d", actualPayload, h.Length)
//	}
//
//	return size
//}
