package leaky_bucket

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	// ErrLimitExhausted is returned by the Limiter in case the number of requests overflows the capacity of a Limiter.
	ErrLimitExhausted = errors.New("requests limit exhausted")
)

type bucket struct {
	mu       sync.Mutex
	capacity int64
	rate     int64
	last     int64 //timestamp của request cuối cùng trong queue
}

func newBucket(capacity int64, rate time.Duration) *bucket {
	return &bucket{
		capacity: capacity,
		rate:     int64(rate),
	}
}

//Ý tưởng: Queue các request vào bucket bằng cách gán cho mỗi request 1 thời gian wait để xử lý
// lần lượt. Request sau wait lâu hơn request trước 1 khoảng thời gian rate.

func (b *bucket) Limit() (time.Duration, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now().UnixNano()
	fmt.Println("Last", b.last, "Now", now)
	if now < b.last {
		// Queue đã có requests: chuyển request hiện tại xuống cuối queue
		b.last += b.rate
	} else {
		// Queue rỗng.

		// offset: trong trường hợp request hiện tại đến trước thời gian rate
		var offset int64
		delta := now - b.last
		if delta < b.rate {
			offset = b.rate - delta
		}
		b.last = now + offset
	}
	wait := b.last - now
	current := int64(math.Ceil(float64(wait)/float64(b.rate)))
	fmt.Println(current, b.capacity)
	if current >= b.capacity {
		b.last = now + b.capacity*b.rate
		return time.Duration(wait), ErrLimitExhausted
	}
	return time.Duration(wait), nil
}
