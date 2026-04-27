package cooldown

import (
	"sync"
	"time"
)

type Tracker struct {
	mu              sync.RWMutex
	lastAlert       map[string]time.Time
	cooldownMinutes int
	resetThreshold  int
	// resetAbove: when true, cooldown resets if fill rises ABOVE resetThreshold
	// (used for empty-bucket alerts). When false, resets if fill drops BELOW
	// resetThreshold (used for full-bucket alerts).
	resetAbove bool
}

func NewTracker(cooldownMinutes, resetThreshold int, resetAbove bool) *Tracker {
	return &Tracker{
		lastAlert:       make(map[string]time.Time),
		cooldownMinutes: cooldownMinutes,
		resetThreshold:  resetThreshold,
		resetAbove:      resetAbove,
	}
}

// IsInCooldown checks if the key is within the cooldown period
func (t *Tracker) IsInCooldown(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	lastTime, ok := t.lastAlert[key]
	if !ok {
		return false
	}

	return time.Since(lastTime) < time.Duration(t.cooldownMinutes)*time.Minute
}

// Record marks the current time as when an alert was sent for the key
func (t *Tracker) Record(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastAlert[key] = time.Now()
}

// Reset removes the cooldown for the key
func (t *Tracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.lastAlert, key)
}

// ShouldReset returns true if the fill percentage has crossed the reset threshold.
// For full-bucket trackers (resetAbove=false): resets when fill drops below threshold.
// For empty-bucket trackers (resetAbove=true): resets when fill rises above threshold.
func (t *Tracker) ShouldReset(fillPercentage float64) bool {
	if t.resetAbove {
		return fillPercentage > float64(t.resetThreshold)
	}
	return fillPercentage < float64(t.resetThreshold)
}
