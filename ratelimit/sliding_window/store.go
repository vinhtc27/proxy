package sliding_window

import "time"

// Store is the interface that represents limiter internal data store.
// Any database struct that implements Store should have functions for incrementing counter of a
// given key and getting counter values of a given key for previous and current window
type store interface {
	// inc increments current window limit counter for key
	inc(key string, window time.Time) error
	// get gets value of previous window counter and current window counter for key
	get(key string, previousWindow, currentWindow time.Time) (prevValue int64, currValue int64, err error)
}
