package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"net"
	"os"

	"golang.org/x/net/ipv6"
)

var be = binary.BigEndian

type reader interface {
	io.Reader
	io.ByteReader
}

type pkt struct {
	reader
	conn *ipv6.PacketConn
	src  net.Addr
}

func process(packet *pkt) {
	msg, _ := packet.ReadByte()

	xid := make([]byte, 3)
	packet.Read(xid)

	var optLen uint16
	binary.Read(packet, be, &optLen)

	switch msg {
	case 1:
		// solicit
		log.Print("replying to ", packet.src)

		// generate address
		ip6addr := []byte{0x2a, 0x02, 0x81, 0x0d, 0xa4, 0x80, 0x2b, 0x00,
			0, 0, 0, 0, 0, 0, 0, 0}
		binary.BigEndian.PutUint64(ip6addr[8:], rand.Uint64())

		// generate reply
		b := &bytes.Buffer{}
		b.WriteByte(7) // reply
		b.Write(xid)

		binary.Write(b, be, uint16(3))     // IA_NA
		binary.Write(b, be, uint16(12+28)) // opt len
		b.Write([]byte{0, 0, 0, 0})        // IAID
		binary.Write(b, be, uint32(180))   // T1
		binary.Write(b, be, uint32(120))   // T2

		binary.Write(b, be, uint16(5))   // IAADDR
		binary.Write(b, be, uint16(24))  // opt len
		b.Write(ip6addr)                 // IPv6-address
		binary.Write(b, be, uint32(180)) // preferred-lifetime
		binary.Write(b, be, uint32(120)) // valid-lifetime

		packet.conn.WriteTo(b.Bytes(), nil, packet.src)
	}
}

func main() {
	pc, err := net.ListenPacket("udp6", "[::]:547")
	chk(err)
	defer pc.Close()

	iface, err := net.InterfaceByName(os.Args[1])
	chk(err)

	grp := &net.UDPAddr{IP: net.ParseIP("ff02::1:2")}

	c := ipv6.NewPacketConn(pc)
	chk(c.JoinGroup(iface, grp))
	defer func() {
		chk(c.LeaveGroup(iface, grp))
	}()

	log.Print("listening")
	for {
		buf := make([]byte, 1024)
		n, _, addr, err := c.ReadFrom(buf)
		chk(err)
		buf = buf[:n]

		go process(&pkt{
			reader: bytes.NewReader(buf),
			conn:   c,
			src:    addr,
		})
	}
}

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
