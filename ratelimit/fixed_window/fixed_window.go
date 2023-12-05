package fixed_window

import (
	"fmt"
	"net/http"
	"time"
)

var requestThrottler = newWindow(3, 100*time.Millisecond)

func RequestThrottler(h http.Handler, _ int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requestThrottler.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			fmt.Println("Rate limit [fixed_window]: too many requests")
			return
		}
		h.ServeHTTP(w, r)
	})
}
