package lb

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

var requestThrottler = newThrottler(1000 * time.Millisecond) 

// ReqThrottledHandler wraps an http.Handler with per host request throttling
// to the specified request maxAmount, responding with 429 when throttled.
// Hiện có thể limit số request vào bucket nhưng chưa drip được request ra theo rate
func RequestThrottler(h http.Handler, capacity int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if wait, err := requestThrottler.Halt(host, capacity); err != nil {
			http.Error(w, "Too many requests [RequestThrottler]", 429)
			return
		} else {
			fmt.Println(wait)
			// time.Sleep(wait)
		}
		h.ServeHTTP(w, r)
	})
}