package sliding_window

import (
	"log"
	"net/http"
	"proxy/utils"
	"time"
)

var (
	windowSize  = time.Second
	windowStore = newLocalStore(2*windowSize, 100*time.Millisecond)
)

func RequestThrottler(h http.Handler, maxAmount int64) http.Handler {
	requestThrottler := newWindow(windowStore, maxAmount, windowSize)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		remoteIP := utils.GetRemoteIP(r)
		//key := fmt.Sprintf("%s_%s_%s", remoteIP, r.URL.String(), r.Method)

		limitStatus, err := requestThrottler.Halt(remoteIP)
		if err != nil {
			// if rate limit error then pass the request
			h.ServeHTTP(w, r)
		}
		if limitStatus.IsLimited {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		if err := requestThrottler.inc(remoteIP); err != nil {
			log.Printf("could not increment key: %s", remoteIP)
		}

		h.ServeHTTP(w, r)
	})
}
