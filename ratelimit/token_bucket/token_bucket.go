package token_bucket

import (
	"net"
	"net/http"
	"time"
)

var requestThrottler = newThrottler(100 * time.Millisecond) // refill every 100ms = 10req/s

// ReqThrottledHandler wraps an http.Handler with per host request throttling
// to the specified request maxAmount, responding with 429 when throttled.
func RequestThrottler(h http.Handler, maxAmount int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if requestThrottler.Halt(host, 1, maxAmount) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}
