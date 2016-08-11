package	main

import (
	"fmt"
	"net/http"
	"time"
)

const	MAX_CONNS	int = 10000

// see these notes on Go's http package's timeout handling defaults:
// - https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779#.fmj7do7v1
// - https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/

func	main() {
	for i := 0; i < MAX_CONNS; i++ {
		go func(id int) {
			var hc = &http.Client{
				Timeout: time.Second * 10,
			}

			response, err := hc.Get("localhost:80")

			if err != nil {
				fmt.Println("err: ", id, err)
			} else {
				defer response.Body.Close()
			}
		}(i)
	}
}
