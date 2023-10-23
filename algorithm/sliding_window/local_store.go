package sw

import (
	"fmt"
	"sync"
	"time"
)

type limitValue struct {
	val        int64
	lastUpdate time.Time
}

// localStore represents internal limiter data database where data are stored in golang maps
type localStore struct {
	data           map[string]limitValue
	mutex          sync.RWMutex
	expirationTime time.Duration
}

// newLocalStore creates new in-memory data store for internal limiter data.
// Each element of localStore is set as expired after expirationTime from its last counter increment.
// Expired elements are removed with a period specified by the flushInterval argument
func newLocalStore(expirationTime time.Duration, flushInterval time.Duration) (m *localStore) {
	m = &localStore{
		data:           make(map[string]limitValue),
		expirationTime: expirationTime,
	}
	go func() {
		ticker := time.NewTicker(flushInterval)
		for range ticker.C {
			m.mutex.Lock()
			for key, val := range m.data {
				if val.lastUpdate.Before(time.Now().UTC().Add(-m.expirationTime)) {
					delete(m.data, key)
				}
			}
			m.mutex.Unlock()
		}
	}()
	return m
}

// Inc increments current window limit counter for key
func (m *localStore) inc(key string, window time.Time) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	data := m.data[mapKey(key, window)]
	data.val++
	data.lastUpdate = time.Now().UTC()
	m.data[mapKey(key, window)] = data
	return nil
}

// Get gets value of previous window counter and current window counter for key
func (m *localStore) get(key string, previousWindow, currentWindow time.Time) (prevValue int64, currValue int64, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	prevValue = m.data[mapKey(key, previousWindow)].val
	currValue = m.data[mapKey(key, currentWindow)].val
	return prevValue, currValue, nil
}

// Size returns current length of data map
func (m *localStore) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.data)
}

func mapKey(key string, window time.Time) string {
	return fmt.Sprintf("%s_%s", key, window.Format(time.RFC3339))
}
