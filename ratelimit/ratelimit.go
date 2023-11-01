package algorithm

import "net/http"

type Limiter interface {
	RequestThrottler(h http.Handler, maxRequests int) http.Handler
}
