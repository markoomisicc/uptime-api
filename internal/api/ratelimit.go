package api

import (
	"sync"
	"time"
)

// cooldownLimiter allows one action per key within the given interval.
// State is kept in memory, so it resets on restart and is not shared
// between multiple server instances.
type cooldownLimiter struct {
	mu       sync.Mutex
	last     map[string]time.Time
	interval time.Duration
}

func newCooldownLimiter(interval time.Duration) *cooldownLimiter {
	return &cooldownLimiter{
		last:     make(map[string]time.Time),
		interval: interval,
	}
}

func (l *cooldownLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	if last, ok := l.last[key]; ok && now.Sub(last) < l.interval {
		return false
	}

	l.last[key] = now

	if len(l.last) > 1000 {
		for k, t := range l.last {
			if now.Sub(t) >= l.interval {
				delete(l.last, k)
			}
		}
	}

	return true
}
