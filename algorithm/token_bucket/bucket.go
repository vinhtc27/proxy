package tb

import (
	"math"
	"sync/atomic"
	"time"
)

// bucket defines a generic lock-free implementation of a Token Bucket.
type bucket struct {
	maxAmount    int64         // the maximum number of tokens the bucket can hold
	refillAmount int64         // the number of tokens to be added on each refill
	refillTime   time.Duration // the amount of time between refills
	tokens       int64         // trigger channel to close the bucket

	close chan struct{} // trigger channel to close the bucket
}

// newbucket returns a full bucket with maxAmount and starts a filling
// go-routine which ticks every refillTime. The number of tokens added on each tick
// is computed dynamically to be even across the duration of a second.
//
// If refillTime == -1 then the filling go-routine won't be started. Otherwise,
// If refillTime < 1/maxAmount seconds, then it will be adjusted to 1/maxAmount seconds.
func newBucket(maxAmount int64, refillTime time.Duration) *bucket {
	b := &bucket{
		tokens:    maxAmount,
		maxAmount: maxAmount,
		close:     make(chan struct{}),
	}

	if evenRefillTime := time.Duration(1e9 / maxAmount); refillTime < evenRefillTime {
		refillTime = evenRefillTime
	}

	b.refillAmount = int64(math.Floor(.5 + (float64(maxAmount) * refillTime.Seconds())))
	b.refillTime = refillTime

	if refillTime == -1 {
		return b
	}

	go b.fill()
	return b
}

// Take attempts to take n tokens out of the bucket.
// This method is thread-safe.
func (b *bucket) Take(n int64) (taken int64) {
	for {
		if tokens := atomic.LoadInt64(&b.tokens); tokens == 0 { // If tokens == 0, nothing will be taken.
			return 0
		} else if n <= tokens {
			if !atomic.CompareAndSwapInt64(&b.tokens, tokens, tokens-n) { // If n <= tokens, n tokens will be taken.
				continue
			}
			return n
		} else if atomic.CompareAndSwapInt64(&b.tokens, tokens, 0) { // If n > tokens, all tokens will be taken.
			return tokens
		}
	}
}

// Put attempts to add n tokens to the bucket.
// This method is thread-safe.
func (b *bucket) Put(n int64) (added int64) {
	for {
		if tokens := atomic.LoadInt64(&b.tokens); tokens == b.maxAmount { // If tokens == capacity, nothing will be added.
			return 0
		} else if left := b.maxAmount - tokens; n <= left {
			if !atomic.CompareAndSwapInt64(&b.tokens, tokens, tokens+n) { // If n <= capacity - tokens, n tokens will be added.
				continue
			}
			return n
		} else if atomic.CompareAndSwapInt64(&b.tokens, tokens, b.maxAmount) { // If n > capacity - tokens, capacity - tokens will be added.
			return left
		}
	}
}

// Wait waits for n amount of tokens to be available.
// If n tokens are immediatelly available it doesn't sleep.
// Otherwise, it sleeps the minimum amount of time required for the remaining
// tokens to be available. It returns the wait duration.
//
// This method is thread-safe.
func (b *bucket) Wait(n int64) time.Duration {
	var rem int64
	if rem = n - b.Take(n); rem == 0 {
		return 0
	}

	var wait time.Duration
	for rem > 0 {
		sleep := b.wait(rem)
		wait += sleep
		time.Sleep(sleep)
		rem -= b.Take(rem)
	}
	return wait
}

// Close stops the filling go-routine given it was started.
func (b *bucket) Close() error {
	close(b.close)
	return nil
}

// wait returns the minimum amount of time required for n tokens to be available.
// if n > maxAmount, n will be adjusted to maxAmount
func (b *bucket) wait(n int64) time.Duration {
	return time.Duration(int64(math.Ceil(math.Min(float64(n), float64(b.maxAmount))/float64(b.refillAmount))) *
		b.refillTime.Nanoseconds())
}

func (b *bucket) fill() {
	ticker := time.NewTicker(b.refillTime)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-b.close:
			return
		default:
			b.Put(b.refillAmount)
		}
	}
}
