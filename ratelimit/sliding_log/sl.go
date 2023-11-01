package sliding_log

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type SlidingLogLimiter struct {
	mu          sync.RWMutex
	hostLog     map[string][]time.Time
	interval    time.Duration
	maxRequests int
}

func NewSlidingLogLimiter(interval time.Duration, maxRequests int) *SlidingLogLimiter {
	return &SlidingLogLimiter{
		hostLog:     make(map[string][]time.Time),
		interval:    interval,
		maxRequests: maxRequests,
	}
}

func (sll *SlidingLogLimiter) Halt(host string) bool {
	sll.mu.Lock()
	defer sll.mu.Unlock()

	now := time.Now()
	_, ok := sll.hostLog[host]
	if !ok {
		sll.hostLog[host] = []time.Time{now}
		return false
	}

	// Đếm số request trong thời gian tính từ thời điểm gọi đến trước đó 1 khoảng thời gian bằng cách duyệt từ cuối hostLog
	requestCount := 0
	timestamp := now.Add(-sll.interval)
	newTs := sll.hostLog[host][:0]
	for i := len(sll.hostLog[host]) - 1; i >= 0; i-- {
		t := sll.hostLog[host][i]
		if !t.Before(timestamp) {
			newTs = append(newTs, t)
			requestCount++
		} else {
			break
		}
	}
	sll.hostLog[host] = newTs

	fmt.Printf("Number of requests within the last %.0f seconds: %d\n", sll.interval.Seconds(), requestCount)
	fmt.Printf("Max request in %.0f seconds: %d\n", sll.interval.Seconds(), sll.maxRequests)

	// Nếu vượt quá số request max thì trả về too many request
	if requestCount >= sll.maxRequests {
		return true
	}

	sll.hostLog[host] = append(sll.hostLog[host], now)
	return false
}

func RequestThrottlerMiddleware(h http.Handler, maxRequests int) http.Handler {
	throttler := NewSlidingLogLimiter(10*time.Second, maxRequests)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if throttler.Halt(host) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}
