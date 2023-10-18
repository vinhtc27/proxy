package tb

import (
	"net"
	"net/http"
	"time"
)

var byteThrottler = newthrottler(25 * time.Millisecond)

// ByteThrottledHandler wraps an http.Handler with per host byte throttling to
// the specified byte rate, responding with 429 when throttled.
func ByteTokenBucket(h http.Handler, rate int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if byteThrottler.Halt(host, r.ContentLength, rate) {
			http.Error(w, "Too many requests [ContentLength]", 429)
			return
		}
		h.ServeHTTP(w, r)
	})
}

var reqThrottler = newthrottler(100 * time.Millisecond)

// ReqThrottledHandler wraps an http.Handler with per host request throttling
// to the specified request rate, responding with 429 when throttled.
func ReqTokenBucket(h http.Handler, rate int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if reqThrottler.Halt(host, 1, rate) {
			http.Error(w, "Too many requests [Rate]", 429)
			return
		}
		h.ServeHTTP(w, r)
	})
}
