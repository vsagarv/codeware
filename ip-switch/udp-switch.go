package	main
// package	udpswitch
// Switch between N input UDP/IP MPEG TS streams, to produce 1 UDP/IP output stream..
// 
//  Switch trigger: read timeout / read error
//  Switch mode: round robin
// 
//  Parameters:
//  - N input IP:PORT pairs and 1 output IP:PORT pair

import	(
	"fmt"
	"log"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"net"
	"time"
)

const (
	maxDatagramSize = 2048	// datagrams are 'usually' smaller ~1500b
	inputReadTimeout = 10 * time.Second // no data for 10seconds on a streaming MPEG TS channel is quite likely a breakdown, with the usual bit rates (VBR/CBR)
)

func	main() {
	srcs := [2]string {"228.0.0.4:5001", "228.0.0.4:5002"}
	dst := "228.0.0.50:6001"

	stream_out(srcs, dst)
}

func	open_udp_conn(saddr string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", saddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	return conn, err
}

func	open_mc_udp_conn(saddr string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", saddr)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)	// TODO: set Interface
	l.SetReadBuffer(maxDatagramSize)

	return l, err
}

// stream_out():
// 
func	stream_out(srcs [2]string, dst string) {
	var	src_conns []*net.UDPConn

	// setup source connections
	for _, src := range srcs {
		conn, err := open_mc_udp_conn(src)
		if err != nil {
			// don't die unless all input sources are DOA
			log.Println(err)
			continue
		}

		src_conns = append(src_conns, conn)
	}

	n_src_conns := len(src_conns)

	log.Println("#sources: ", n_src_conns)

	if n_src_conns == 0 {
		fmt.Println("ERROR: No input sources are reachable")
		return
	}

	// setup destination connection
	dst_conn, err := open_udp_conn(dst)
	if err != nil {
		fmt.Println("ERROR: Output destination unreachable")
		return
	}

	// catch SIGINT/TERM and setup a janitor
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		for _, c := range src_conns {
			c.Close()
		}
		dst_conn.Close()

		log.Fatal("bailing out on SIGINT/TERM ...")
	}()
	// end of setup

	curr_conn_idx := 0
	next_conn_idx := 0

	for {
		var buf [maxDatagramSize]byte

		conn := src_conns[next_conn_idx]
		conn.SetReadDeadline(time.Now().Add(inputReadTimeout))

		// n, err := src_conns[next_conn].Read(buf[0:])
		n, _, err := conn.ReadFromUDP(buf[0:])

		// switch to the next connection on error
		if err != nil {
			log.Println(err)

			curr_conn_idx = next_conn_idx
			next_conn_idx = (next_conn_idx + 1) % n_src_conns

			log.Println("switching from connection#" + strconv.Itoa(curr_conn_idx) + " to connection#" + strconv.Itoa(next_conn_idx))
		} else {
			_, err = dst_conn.Write(buf[0:n])
			if err != nil {
				fmt.Println("ERROR: ", err)
				// nothing much can be done; carry on, Jeeves.
			}
		}
	}
}
