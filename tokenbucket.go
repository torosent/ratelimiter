package tokenbucket

import (
	"sync"
	"time"
)

// TokenBucket ...
type TokenBucket struct {
	rps             int64
	nextRefillTime  int64
	availableTokens int64
	capacity        int64
	takeTokens      int64
	mux             sync.Mutex
}

//New ...
func New(fillInterval time.Duration, capacity int64) *TokenBucket {
	rate := int64(fillInterval)
	tb := TokenBucket{
		rps:        rate,
		capacity:   capacity,
		takeTokens: 1,
	}
	return &tb
}

//NewWithRate for example: 100 requests per second
func NewWithRate(rps int) *TokenBucket {
	rate := int64(time.Second / time.Duration(rps))
	tb := TokenBucket{
		rps:        rate,
		capacity:   1,
		takeTokens: 1,
	}
	return &tb
}

func (t *TokenBucket) take() {

	refillAmount := t.refill()
	newTokens := min(t.capacity, max(0, refillAmount))
	t.availableTokens = max(0, min(t.availableTokens+newTokens, t.capacity))
	if t.takeTokens > t.availableTokens {
		return
	}
	t.availableTokens -= t.takeTokens

}

func (t *TokenBucket) refill() int64 {
	now := time.Now().UnixNano()
	if now < t.nextRefillTime {
		return 0
	}

	refillAmount := max((now-t.nextRefillTime)/t.rps, 1)

	t.nextRefillTime += t.rps * refillAmount
	return t.capacity * refillAmount
}

func (t *TokenBucket) consume() time.Duration {

	var duration time.Duration
	now := time.Now().UnixNano()
	if t.nextRefillTime == 0 {
		t.nextRefillTime = now
	}

	if t.availableTokens == 0 {
		duration = time.Duration(t.nextRefillTime - now)
	}
	return duration
}

//ConsumeWithoutBlock ...
func (t *TokenBucket) ConsumeWithoutBlock() time.Duration {
	t.mux.Lock()
	defer t.mux.Unlock()

	duration := t.consume()
	t.take()
	return duration
}

// ConsumeWithBlock ...
func (t *TokenBucket) ConsumeWithBlock() time.Time {
	t.mux.Lock()
	defer t.mux.Unlock()

	duration := t.consume()

	if t.availableTokens == 0 {
		time.Sleep(duration)
	}

	t.take()
	return time.Now()
}
