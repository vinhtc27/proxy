package fw

import (
	"net/http"
	"time"
)

var requestThrottler = newWindow(3, 100*time.Millisecond)

func RequestThrottler(h http.Handler, _ int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requestThrottler.Allow() {
			http.Error(w, "Reject", http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}