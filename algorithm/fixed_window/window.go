package fixed_window

import (
	"sync"
	"time"
)

type window struct {
	limit           int
	windowSize      time.Duration
	requestCount    int
	lastRequestTime time.Time
	mu              sync.Mutex
}

func newWindow(limit int, windowSize time.Duration) *window {
	return &window{
		limit:           limit,
		windowSize:      windowSize,
		requestCount:    0,
		lastRequestTime: time.Now(),
	}
}

func (w *window) Allow() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	if now.Sub(w.lastRequestTime) > w.windowSize {
		w.lastRequestTime = now
		w.requestCount = 1
		return true
	}
	if w.requestCount < w.limit {
		w.requestCount++
		return true
	}
	return false
}
