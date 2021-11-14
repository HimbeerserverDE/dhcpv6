package main

import (
	"log"
	"net"
	"os"

	"golang.org/x/net/ipv6"
)

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

		log.Printf("buf[%d]: %v from addr %v\n", n, buf, addr)
	}
}

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
