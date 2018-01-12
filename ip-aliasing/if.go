// source: https://play.golang.org/p/E3iXXpq0mD
// see also: http://stackoverflow.com/questions/27410764/dial-with-a-specific-address-interface-golang

package main

import (
	"log"
	"net"
)

func main() {
	ifaces, err := net.Interfaces()
	ep(err)
	for _, iface := range ifaces {
		if iface.Name == "eth0" || iface.Name == "en0" {
			addrs, err := iface.Addrs()
			ep(err)
			addr, ok := addrs[0].(*net.IPNet)
			if !ok {
				log.Fatal("Address is not an IP Address:", addrs[0])
			}
			log.Println("Found:", addr.IP)

		}
	}
}

func ep(err error) error {
	if err != nil {
		log.Panic(err)
	}
	return err
}
