package lb

import (
	"sync"
	"time"
)

type throttler struct {
	mutex	sync.RWMutex       // mutex to protect the buckets map
	rate	time.Duration      // the amount of time between refills of each bucket
	buckets	map[string]*bucket // the map of buckets
	// close	chan struct{}      // trigger channel to close the throttler
}


func newThrottler(rate time.Duration) *throttler {
	th := &throttler{
		rate: rate,
		buckets:    map[string]*bucket{},
		// close:      make(chan struct{}),
	}
	return th
}

func (t *throttler) Bucket(key string, capacity int64) *bucket {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	b, ok := t.buckets[key]

	if !ok {
		b = newBucket(capacity, t.rate)
		t.buckets[key] = b
	}

	return b
}

func (t *throttler) Halt(key string, capacity int64) (time.Duration, error) {
	b := t.Bucket(key, capacity)

	if wait, err := b.Limit(); err == nil {
		return wait, nil
	} else {
		return wait, err
	}
}