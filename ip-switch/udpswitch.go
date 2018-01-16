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
	"log"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"net"
	"time"
)

const (
	maxDatagramSize = 8192	// datagrams are 'usually' far smaller ~1500b
	inputReadTimeout = 10 * time.Second // no data for 10seconds on a streaming MPEG TS channel is quite likely a breakdown, with the usual bit rates (VBR/CBR)
)

func	open_udp_conn(saddr string) (*net.UDPConn, error) {

	addr, err := net.ResolveUDPAddr("udp4", saddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadBuffer(maxDatagramSize)

	return conn, err
}

func	listen_udp_conn(saddr string) (*net.UDPConn, error) {

	addr, err := net.ResolveUDPAddr("udp", saddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadBuffer(maxDatagramSize)

	return conn, err
}

// TODO: DRY with listen_udp_conn
func	listen_mc_udp_conn(saddr string) (*net.UDPConn, error) {

	addr, err := net.ResolveUDPAddr("udp", saddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)	// TODO: set Interface
	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadBuffer(maxDatagramSize)

	return conn, err
}

// stream_out():
// 
func	stream_out(srcs [2]string, dst string) {
	var	src_conns []*net.UDPConn

	// listen to sources
	for _, src := range srcs {
		// conn, err := listen_mc_udp_conn(src)
		conn, err := listen_udp_conn(src)
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
		log.Println("ERROR: No input sources are reachable")
		return
	}

	// setup destination connection
	dst_conn, err := open_udp_conn(dst)
	if err != nil {
		log.Println("ERROR: Output destination unreachable")
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
				log.Println("ERROR: ", err)
				// nothing much can be done; carry on, Jeeves.
			}
		}
	}
}


// udp-switch test driver

func	main() {
	// unicast
	srcs := [2]string {"127.0.0.1:15001", "127.0.0.1:15002"}
	dst := "127.0.0.1:16001"

	// multicast
	// srcs := [2]string {"228.0.0.4:15001", "228.0.0.4:15002"}
	// dst := "228.0.0.50:16001"

	stream_out(srcs, dst)
}
