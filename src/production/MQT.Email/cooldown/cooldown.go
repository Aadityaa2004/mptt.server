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
}

func NewTracker(cooldownMinutes, resetThreshold int) *Tracker {
	return &Tracker{
		lastAlert:       make(map[string]time.Time),
		cooldownMinutes: cooldownMinutes,
		resetThreshold:  resetThreshold,
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

// ShouldReset returns true if the fill percentage is below the reset threshold
func (t *Tracker) ShouldReset(fillPercentage float64) bool {
	return fillPercentage < float64(t.resetThreshold)
}
