package sliding_log

import (
	"fmt"
	"sync"
	"time"
)

type SlidingLogLimiter struct {
	mu       sync.RWMutex
	hostLog  map[string][]time.Time
	interval time.Duration
}

func newSlidingLogLimiter(interval time.Duration) *SlidingLogLimiter {
	return &SlidingLogLimiter{
		hostLog:  make(map[string][]time.Time),
		interval: interval,
	}
}

func (sll *SlidingLogLimiter) Halt(host string, maxRequests int) bool {
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
	fmt.Printf("Max request in %.0f seconds: %d\n", sll.interval.Seconds(), maxRequests)

	// Nếu vượt quá số request max thì trả về too many request
	if requestCount >= maxRequests {
		return true
	}

	sll.hostLog[host] = append(sll.hostLog[host], now)
	return false
}
