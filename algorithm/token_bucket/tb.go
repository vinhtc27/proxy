package tb

import (
	"net"
	"net/http"
	"time"
)

var byteThrottler = newThrottler(25 * time.Millisecond) // refill every 25ms = 40KB/s

// ByteThrottledHandler wraps an http.Handler with per host byte throttling to
// the specified byte maxAmount, responding with 429 when throttled.
func ByteThrottler(h http.Handler, maxAmount int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if byteThrottler.Halt(host, r.ContentLength, maxAmount) {
			http.Error(w, "Too many requests [ByteThrottler]", 429)
			return
		}
		h.ServeHTTP(w, r)
	})
}

var requestThrottler = newThrottler(100 * time.Millisecond) // refill every 100ms = 10req/s

// ReqThrottledHandler wraps an http.Handler with per host request throttling
// to the specified request maxAmount, responding with 429 when throttled.
func RequestThrottler(h http.Handler, maxAmount int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if requestThrottler.Halt(host, 1, maxAmount) {
			http.Error(w, "Too many requests [RequestThrottler]", 429)
			return
		}
		h.ServeHTTP(w, r)
	})
}
