// This is a golang redigo based redis client that can be used to open (many)
// concurrent connections to a redis server, with socket lingering off. Due to
// this, the client gets to close all open concurrent connections as soon as
// possible, without the sockets entering TIME_WAIT state.
//
// In order to achieve this sort of concurrency, the following changes are
// suggested on the client side:
//
// ==== increase soft & hard limits on open FDs for root & normal users ===
// ubuntu@redigo-client:~$ tail -7 /etc/security/limits.conf
//
// *    soft nofile 64000
// *    hard nofile 64000
// root soft nofile 64000
// root hard nofile 64000
//
// # End of file
// ====
//
// On the (redis) server increase FDs as above and tune networking as follows:
// ====
// // ubuntu@redis-ports:~$ tail -27 /etc/sysctl.conf
//
// # Allow reuse of sockets in TIME_WAIT state for new connections
// # only when it is safe from the network stack’s perspective.
// net.ipv4.tcp_tw_reuse = 1
//
// # pending connections are kept in a socket buffer; 500K per socket
// #
// net.core.rmem_max = 500000
// net.core.wmem_max = 500000
//
// # Increase the number of outstanding syn requests allowed.
// # c.f. The use of syncookies.
// net.ipv4.tcp_max_syn_backlog = 10000
// net.ipv4.tcp_syncookies = 1
//
// # The maximum number of "backlogged sockets".  Default is 128.
// net.core.somaxconn = 10000
//
// # How may times to retry before killing TCP connection, closed by our side.
// # Avoids too many sockets in FIN-WAIT-1 state (default 0 which means 8!).
// # (see also: /proc/sys/net/ipv4/tcp_max_orphans)
// net.ipv4.tcp_orphan_retries = 1
//
// # Time to hold socket in state FIN-WAIT-2, if it was closed by our side
// # Reduces time for sockets to be in FIN-WAIT-2 state (default 60secs).
// net.ipv4.tcp_fin_timeout = 30
// #
// ubuntu@redis-ports:~$
//
// And run 'sysctl -p' after modifying sysctl.conf as above
// ====

package	main

import	"flag"
import	"fmt"
import	"time"
import	"net"
import	"strconv"
import  "sync"
import	"github.com/garyburd/redigo/redis"

// This is one explanation of turning off socket lingering. This depends
// heavily on the OS flavour.:
//
// When socket lingering is off, close() returns immediately. The underlying
// stack discards any unsent data, and, in the case of connection-oriented
// protocols such as TCP, sends a RST (reset) to the peer (this is termed
// a hard or abortive close). All subsequent attempts by the peer's application
// to read()/recv() data will result in an ECONNRESET.

const	SO_LINGER_OFF	int = 0

const	RedisAddr	string = "172.31.2.145:6379"

// setupRedisPool(): Custom dialer function to return a TCP connection
// with socketing lingering turned off.
func	setupRedisPool(so_linger int) (redis.Conn, error) {
	redisAddr, _ := net.ResolveTCPAddr("tcp", RedisAddr)

	tc, err := net.DialTCP("tcp", nil, redisAddr)

	if err != nil {
		fmt.Println("setupRedisPool: net.DialTCP: ", err)
	}

	if err := tc.SetKeepAlive(false); err != nil {
		fmt.Println("setupRedisPool: tc.SetKeepAlive: ", err)
	}

	if so_linger == SO_LINGER_OFF {
		if err := tc.SetLinger(0); err != nil {
			fmt.Println("setupRedisPool: tc.SetLinger: ", err)
		}
	}

	c := redis.NewConn(tc, 10*time.Second, 10*time.Second)

	return c, nil
}

func	newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle: 9000,
		MaxActive: 9000,
		IdleTimeout: 300 * time.Second,
		Dial: func () (redis.Conn, error) {
			return setupRedisPool(SO_LINGER_OFF)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// runRedisClients():
// Launches multiple concurrent client connection bursts,
// with 'cps' concurrent client connections per burst,
// each burst seperated by 'gaps' seconds from the next burst,
// and the total run spread *approximately* over 'durs' seconds.
func	runRedisClients(pool *redis.Pool, cps, durs, gaps int) {
	epoch := time.Now()
	nBursts := durs/gaps
	wgs := make([]sync.WaitGroup, nBursts)

	c := pool.Get()
	defer c.Close()

	c.Do("SET", "foo", "-1:-1")

	for i := 0; i < nBursts; i++ {
		msgBuf := make([]string, cps)	// one debug msg per client

		wgs[i].Add(cps)
		for j := 0; j < cps; j++ {
			go func(id, cid int) {
				defer wgs[id].Done()

				conn := pool.Get()
				defer conn.Close()

				if _, cErr := conn.Do("PING"); cErr != nil {
					fmt.Println("Bad connection from pool: ", cErr)
					return
				}

				// "foo" might be set in the previous burst and read by a large number of go()s in this burst!!
				v, err := redis.String(conn.Do("GET", "foo"))

				msgBuf[cid] = "@" + time.Since(epoch).String() + " - GET foo:" + v

				s, err := conn.Do("SET", "foo", strconv.Itoa(id) + ":" + strconv.Itoa(cid))
				conn.Flush()

				// "foo" will be overwritten 'cps' times and finally be set to the goroutine that executes last, in this burst

				if err != nil {
					fmt.Println("SET foo: ", s, err)
				}

				msgBuf[cid] += " - " + time.Since(epoch).String()
			}(i, j)
		}

		wgs[i].Wait() // wait for the above client burst to complete

		fmt.Println("pool ActiveCount: ", pool.ActiveCount())

		// dump the timestamped messages of the burst
		fmt.Println("====", i, "====")
		for j := 0; j < cps; j++ {
			fmt.Println(msgBuf[j])
		}

		time.Sleep(time.Duration(gaps) * time.Second)	// gap between bursts
	}
}

const	CPS int = 5	// # concurrent clients per second
const	DURS int = 60	// # duration in seconds, of the entire run
const	GAPS int = 5	// # gap in seconds between concurrent bursts

var	pool *redis.Pool

func	main() {
	cpsPtr := flag.Int("cps", CPS, "# concurrent clients per second")
	dursPtr := flag.Int("durs", DURS, "# duration in seconds, of the entire run")
	gapsPtr := flag.Int("gaps", GAPS, "# gap in seconds between concurrent bursts")

	flag.Parse()

	fmt.Println("cps, durs, gaps:", *cpsPtr, *dursPtr, *gapsPtr)

	if *cpsPtr != 0 {
		pool = newPool()
		runRedisClients(pool, *cpsPtr, *dursPtr, *gapsPtr)
	}

	fmt.Println("arbitrarily waiting for 10 seconds to let goroutines exit!! ...")
	time.Sleep(10)

	fmt.Println("closing the redigo pool and exiting.")

	if pool != nil {
		fmt.Println("pool ActiveCount: ", pool.ActiveCount())

		// IdleCount() is a custom implementation; see comments below
		// fmt.Println("pool IdleCount: ", pool.IdleCount())

		pool.Close()

		fmt.Println("pool ActiveCount: ", pool.ActiveCount())
		// fmt.Println("pool IdleCount: ", pool.IdleCount())
	}
}

// Add this to "github.com/garyburd/redigo/redis/pool.go" for IdleCount()
//
// IdleCount returns the number of active connections in the pool.
// func (p *Pool) IdleCount() int {
//        p.mu.Lock()
//        idles := p.idle.Len()
//        p.mu.Unlock()
//        return idles
// }
