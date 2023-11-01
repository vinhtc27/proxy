package sliding_window

import (
	"time"
)

// simple rate-limiter for any resources inspired by Cloudflare's approach: https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
type window struct {
	windowStore store
	windowSize  time.Duration
	maxAmount   int64
}

func newWindow(windowStore store, maxAmount int64, windowSize time.Duration) *window {
	return &window{
		windowStore: windowStore,
		maxAmount:   maxAmount,
		windowSize:  windowSize,
	}
}

// Inc increments limiter counter for a given key or returns error when it's not possible
func (w *window) inc(key string) error {
	currentSize := time.Now().UTC().Truncate(w.windowSize)
	return w.windowStore.inc(key, currentSize)
}

// LimitStatus represents current status of limitation for a given key
type LimitStatus struct {
	// IsLimited is true when a given key should be rate-limited
	IsLimited bool
	// LimitDuration is not nil when IsLimited is true. It's the time for which a given key should be blocked before CurrentRate falls below declared in constructor requests limit
	LimitDuration *time.Duration
	// CurrentRate is approximated current requests rate per window size (declared in the constructor)
	CurrentRate float64
}

// Halt checks status of rate-limiting for a key. It returns error when limiter data could not be read
func (w *window) Halt(key string) (limitStatus *LimitStatus, err error) {
	currentSize := time.Now().UTC().Truncate(w.windowSize)
	previousSize := currentSize.Add(-w.windowSize)
	prevValue, currentValue, err := w.windowStore.get(key, previousSize, currentSize)
	if err != nil {
		return nil, err
	}
	timeFromCurrWindow := time.Now().UTC().Sub(currentSize)

	rate := float64(
		(float64(w.windowSize)-float64(timeFromCurrWindow))/
			float64(w.windowSize))*float64(prevValue) + float64(currentValue)

	limitStatus = &LimitStatus{}
	if rate >= float64(w.maxAmount) {
		limitStatus.IsLimited = true
		limitDuration := w.calcLimitDuration(prevValue, currentValue, timeFromCurrWindow)
		limitStatus.LimitDuration = &limitDuration
	}
	limitStatus.CurrentRate = rate

	return limitStatus, nil
}

func (r *window) calcRate(timeFromCurrWindow time.Duration, prevValue int64, currentValue int64) float64 {
	return float64((float64(r.windowSize)-float64(timeFromCurrWindow))/float64(r.windowSize))*float64(prevValue) + float64(currentValue)
}

func (r *window) calcLimitDuration(prevValue, currValue int64, timeFromCurrWindow time.Duration) time.Duration {
	// we should find x parameter in equation: x*prevValue+currentValue = r.maxAmount
	// then (1.0-x)*windowSize is duration from current window start when limit can be removed
	// then ((1.0-x)*windowSize) - timeFromCurrWindow is duration since current time to the time when limit can be removed = limitDuration
	// --
	// if prevValue is zero then unblock is in the next window so we should use equation x*currentValue+nextWindowValue = r.maxAmount
	// to calculate x parameter
	var limitDuration time.Duration
	if prevValue == 0 {
		// unblock in the next window where prevValue is currValue and currValue is zero (assuming that since limit start all requests are blocked)
		if currValue != 0 {
			nextWindowUnblockPoint := float64(r.windowSize) * (1.0 - (float64(r.maxAmount) / float64(currValue)))
			timeToNextWindow := r.windowSize - timeFromCurrWindow
			limitDuration = timeToNextWindow + time.Duration(int64(nextWindowUnblockPoint)+1)
		} else {
			// when maxAmount is 0 we want to block all requests - set limitDuration to -1
			limitDuration = -1
		}
	} else {
		currWindowUnblockPoint := float64(r.windowSize) * (1.0 - (float64(r.maxAmount-currValue) / float64(prevValue)))
		limitDuration = time.Duration(int64(currWindowUnblockPoint+1)) - timeFromCurrWindow

	}
	return limitDuration
}
