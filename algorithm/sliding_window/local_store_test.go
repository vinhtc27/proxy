package sliding_window

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLocalStore(t *testing.T) {
	type args struct {
		expirationTime time.Duration
		flushInterval  time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_newLocalStore",
			args: args{
				expirationTime: 1 * time.Minute,
				flushInterval:  2 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		localStore := newLocalStore(tt.args.expirationTime, tt.args.flushInterval)
		assert.Equal(t, tt.args.expirationTime, localStore.expirationTime)
	}
}

func TestLocalStore_Inc(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		window  time.Time
		wantErr bool
	}{
		{
			name:    "test_localStore_Inc",
			key:     "tt",
			window:  time.Now().UTC(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		m := newLocalStore(1*time.Minute, 10*time.Second)
		err := m.inc(tt.key, tt.window)
		assert.NoError(t, err)
		prevVal, currVal, err := m.get(tt.key, tt.window, tt.window)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), prevVal)
		assert.Equal(t, int64(1), currVal)
	}
}

func TestLocalStore_Get(t *testing.T) {
	type args struct {
		key            string
		previousWindow time.Time
		currentWindow  time.Time
	}
	tests := []struct {
		name          string
		args          args
		wantPrevValue int64
		wantCurrValue int64
		wantErr       bool
	}{
		{
			name: "test_localStore_Get",
			args: args{
				key:            "tt",
				previousWindow: time.Now().UTC().Add(-1 * time.Minute),
				currentWindow:  time.Now().UTC(),
			},
			wantPrevValue: 10,
			wantCurrValue: 5,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		m := newLocalStore(1*time.Minute, 10*time.Second)
		m.data[mapKey(tt.args.key, tt.args.previousWindow)] = limitValue{val: tt.wantPrevValue}
		m.data[mapKey(tt.args.key, tt.args.currentWindow)] = limitValue{val: tt.wantCurrValue}

		prevVal, currVal, err := m.get(tt.args.key, tt.args.previousWindow, tt.args.currentWindow)
		assert.NoError(t, err)
		assert.Equal(t, tt.wantPrevValue, prevVal)
		assert.Equal(t, tt.wantCurrValue, currVal)
	}
}

func TestLocalStore_Size(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		window time.Time
		size   int
	}{
		{
			name:   "test_localStore_Size",
			key:    "tt",
			window: time.Now().UTC(),
			size:   1,
		},
		{
			name:   "test_localStore_Size",
			key:    "tt",
			window: time.Time{},
			size:   0,
		},
	}
	for _, tt := range tests {
		m := newLocalStore(1*time.Minute, 10*time.Second)
		if !tt.window.IsZero() {
			err := m.inc(tt.key, tt.window)
			assert.NoError(t, err)
		}
		assert.Equal(t, tt.size, m.Size())
	}
}
