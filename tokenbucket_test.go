package tokenbucket

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var count int32
var wg sync.WaitGroup

var testRatePlans = []struct {
	rps           int
	totalRequests int
	expectedWait  time.Duration
	expectedDelta time.Duration
}{{
	rps:           1,
	expectedDelta: 10 * time.Millisecond,
	expectedWait:  1 * time.Second,
	totalRequests: 10,
},
	{
		rps:           10,
		expectedDelta: 5 * time.Millisecond,
		expectedWait:  100 * time.Millisecond,
		totalRequests: 100,
	}}

var testPlans = []struct {
	capacity      int64
	fillInterval  time.Duration
	totalRequests int
	expectedWait  time.Duration
	expectedDelta time.Duration
}{{
	capacity:      10,
	expectedDelta: 10 * time.Millisecond,
	expectedWait:  1 * time.Second,
	totalRequests: 100,
}}

func TestBucket(t *testing.T) {

	for _, plan := range testPlans {

		tb := New(plan.expectedWait, plan.capacity)

		prev := time.Now()
		for i := 0; i < plan.totalRequests; i++ {
			now := tb.ConsumeWithBlock()
			duration := now.Sub(prev)
			prev = now
			if i > 0 && i%int(plan.capacity) == 0 {
				assert.InDelta(t, plan.expectedWait, duration, float64(plan.expectedDelta))
			}
		}

		assert.EqualValues(t, 0, tb.availableTokens, fmt.Sprintf("expected 0 tokens, got %d", tb.availableTokens))
	}
}

func TestBucketWithRate(t *testing.T) {

	for _, plan := range testRatePlans {

		tb := NewWithRate(plan.rps)

		prev := time.Now()
		for i := 0; i < plan.totalRequests; i++ {
			now := tb.ConsumeWithBlock()
			duration := now.Sub(prev)
			prev = now
			if i > 0 {
				assert.InDelta(t, plan.expectedWait, duration, float64(plan.expectedDelta))
			}
		}

		assert.EqualValues(t, 0, tb.availableTokens, fmt.Sprintf("expected 0 tokens, got %d", tb.availableTokens))
	}
}

func TestConcurrentBucketWithRate(t *testing.T) {

	tasks := 100
	tb := NewWithRate(10)
	count = 0
	wg.Add(tasks)
	defer wg.Wait()

	for i := 0; i < tasks; i++ {
		go job(tb)
	}
	assert.EqualValues(t, 0, tb.availableTokens, fmt.Sprintf("expected 0 tokens, got %d", tb.availableTokens))
}

func job(tb *TokenBucket) {
	defer wg.Done()
	now := tb.ConsumeWithBlock()
	fmt.Println(now)
	atomic.AddInt32(&count, 1)

}
