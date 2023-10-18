package tb

import (
	"sync"
	"time"
)

// throttler is a thread-safe wrapper around a map of buckets and an easy to
// use API for generic throttling.
type throttler struct {
	mutex      sync.RWMutex       // mutex to protect the buckets map
	refillTime time.Duration      // the amount of time between refills of each bucket
	buckets    map[string]*bucket // the map of buckets
	close      chan struct{}      // trigger channel to close the throttler
}

// Newthrottler returns a throttler with a single filler go-routine for all
// its Buckets which ticks every refillTime.
// The number of tokens added on each tick for each bucket is computed
// dynamically to be even accross the duration of a second.
//
// If refillTime <= 0, the filling go-routine won't be started.
func newThrottler(refillTime time.Duration) *throttler {
	th := &throttler{
		refillTime: refillTime,
		buckets:    map[string]*bucket{},
		close:      make(chan struct{}),
	}

	if refillTime > 0 {
		go th.fill()
	}

	return th
}

// Bucket returns a Bucket with maxAmount capacity, keyed by key.
//
// If a Bucket (key, maxAmount) doesn't exist yet, it is created.
//
// You must call Close when you're done with the throttler in order to not leak
// a go-routine and a system-timer.
func (t *throttler) Bucket(key string, maxAmount int64) *bucket {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	b, ok := t.buckets[key]

	if !ok {
		// -1 param means no filling go-routine for this bucket
		// because it's already handled by the throttler's single filling go-routine
		b = newBucket(maxAmount, -1)
		t.buckets[key] = b
	}

	return b
}

// Wait waits for n amount of tokens to be available.
// If n tokens are immediatelly available it doesn't sleep. Otherwise, it sleeps
// the minimum amount of time required for the remaining tokens to be available.
// It returns the wait duration.
//
// If a Bucket (key, maxAmount) doesn't exist yet, it is created.
// If refillTime < 1/maxAmount seconds, the effective wait maxAmount won't be correct.
//
// You must call Close when you're done with the throttler in order to not leak
// a go-routine and a system-timer.
func (t *throttler) Wait(key string, n, maxAmount int64) time.Duration {
	return t.Bucket(key, maxAmount).Wait(n)
}

// Halt returns a bool indicating if the Bucket identified by key and maxAmount has
// n amount of tokens. If it doesn't, the taken tokens are added back to the
// bucket.
//
// If a Bucket (key, maxAmount) doesn't exist yet, it is created.
// If refillTime < 1/maxAmount seconds, the results won't be correct.
//
// You must call Close when you're done with the throttler in order to not leak
// a go-routine and a system-timer.
func (t *throttler) Halt(key string, n, maxAmount int64) bool {
	b := t.Bucket(key, maxAmount)

	if got := b.Take(n); got != n {
		b.Put(got)
		return true
	}

	return false
}

// Close stops filling the Buckets, closing the filling go-routine.
func (t *throttler) Close() error {
	close(t.close)

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, b := range t.buckets {
		b.Close()
	}

	return nil
}

func (t *throttler) fill() {
	ticker := time.NewTicker(t.refillTime)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-t.close:
			return
		default:
		}
		t.mutex.RLock()
		for _, b := range t.buckets {
			b.Put(b.refillAmount)
		}
		t.mutex.RUnlock()
	}
}
