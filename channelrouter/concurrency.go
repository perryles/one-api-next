package channelrouter

import (
	"sync"
	"sync/atomic"
)

type ConcurrencyTracker struct {
	activeRequests sync.Map
}

func NewConcurrencyTracker() *ConcurrencyTracker {
	return &ConcurrencyTracker{}
}

func (t *ConcurrencyTracker) TryAcquire(channelId int, maxConcurrency int) bool {
	if maxConcurrency <= 0 {
		return true
	}
	counter := t.getCounter(channelId)
	current := counter.Load()
	if int(current) >= maxConcurrency {
		return false
	}
	if !counter.CompareAndSwap(current, current+1) {
		current = counter.Load()
		if int(current) >= maxConcurrency {
			return false
		}
		if !counter.CompareAndSwap(current, current+1) {
			return false
		}
	}
	return true
}

func (t *ConcurrencyTracker) Release(channelId int) {
	counter := t.getCounter(channelId)
	newVal := counter.Add(-1)
	if newVal < 0 {
		counter.Store(0)
	}
}

func (t *ConcurrencyTracker) GetActiveCount(channelId int) int64 {
	counter := t.getCounter(channelId)
	return counter.Load()
}

func (t *ConcurrencyTracker) IsAtCapacity(channelId int, maxConcurrency int) bool {
	if maxConcurrency <= 0 {
		return false
	}
	return int(t.GetActiveCount(channelId)) >= maxConcurrency
}

func (t *ConcurrencyTracker) getCounter(channelId int) *atomic.Int64 {
	val, ok := t.activeRequests.Load(channelId)
	if !ok {
		newCounter := &atomic.Int64{}
		actual, loaded := t.activeRequests.LoadOrStore(channelId, newCounter)
		if loaded {
			return actual.(*atomic.Int64)
		}
		return newCounter
	}
	return val.(*atomic.Int64)
}
