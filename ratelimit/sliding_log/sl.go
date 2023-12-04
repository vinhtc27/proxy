package sliding_log

import (
	"net"
	"net/http"
	"time"
)

var throttler = newSlidingLogLimiter(10 * time.Second)

func RequestThrottler(h http.Handler, maxRequests int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if throttler.Halt(host, maxRequests) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}
