// Multicast Receiver - https://gist.github.com/fiorix/9664255
//
// Used a destination (test receiver) for the multicast udp switch output

package main

import (
	"encoding/hex"
	"log"
	"net"
	"time"
)

const (
	udpSrvAddr         = "127.0.0.1:16001"
	udpMCSrvAddr         = "228.0.0.50:16001"
	maxDatagramSize = 8192
)

func main() {
	// go ping(udpMCSrvAddr)	// to test connectivity to the server
	// serveMulticastUDP(udpMCSrvAddr, msgHandler)
	serveUDP(udpSrvAddr, msgHandler)
}

func ping(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	for {
		c.Write([]byte("hello, world\n"))
		time.Sleep(1 * time.Second)
	}
}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	log.Println(n, "bytes read from", src)
	log.Println(hex.Dump(b[:n]))
}

func serveMulticastUDP(a string, h func(*net.UDPAddr, int, []byte)) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	l.SetReadBuffer(maxDatagramSize)
	for {
		b := make([]byte, maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		h(src, n, b)
	}
}

func serveUDP(a string, h func(*net.UDPAddr, int, []byte)) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenUDP("udp", addr)
	l.SetReadBuffer(maxDatagramSize)
	for {
		b := make([]byte, maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		h(src, n, b)
	}
}
