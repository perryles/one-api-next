package channelrouter

import (
	"sync"
	"time"
)

type RPMTracker struct {
	mu      sync.Mutex
	windows map[int]*rpmWindow
}

type rpmWindow struct {
	counts    []int64
	startTime time.Time
	windowSec int
}

func NewRPMTracker() *RPMTracker {
	return &RPMTracker{
		windows: make(map[int]*rpmWindow),
	}
}

func (t *RPMTracker) Increment(channelId int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	w, ok := t.windows[channelId]
	if !ok {
		w = &rpmWindow{
			counts:    make([]int64, 60),
			startTime: time.Now().Truncate(time.Second),
			windowSec: 60,
		}
		t.windows[channelId] = w
	}

	t.advance(w)
	idx := int(time.Now().Unix() - w.startTime.Unix())
	if idx >= 0 && idx < 60 {
		w.counts[idx]++
	}
}

func (t *RPMTracker) CurrentRPM(channelId int) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	w, ok := t.windows[channelId]
	if !ok {
		return 0
	}

	t.advance(w)
	var total int64
	now := time.Now().Unix()
	start := w.startTime.Unix()
	for i := 0; i < 60; i++ {
		ts := start + int64(i)
		if ts > now {
			break
		}
		total += w.counts[i]
	}
	return int(total)
}

func (t *RPMTracker) advance(w *rpmWindow) {
	now := time.Now().Truncate(time.Second)
	elapsed := int(now.Unix() - w.startTime.Unix())
	if elapsed <= 0 {
		return
	}
	if elapsed >= 60 {
		for i := range w.counts {
			w.counts[i] = 0
		}
		w.startTime = now
		return
	}
	for i := 0; i < elapsed; i++ {
		w.counts[i] = 0
	}
	newCounts := make([]int64, 60)
	copy(newCounts, w.counts[elapsed:])
	w.counts = newCounts
	w.startTime = now
}

func (t *RPMTracker) Cleanup(channelId int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.windows, channelId)
}
