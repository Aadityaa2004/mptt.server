package cooldown

import (
	"testing"
	"time"
)

func TestTracker_IsInCooldown_Empty(t *testing.T) {
	tracker := NewTracker(60, 70, false) // 60 min cooldown, 70% reset threshold
	if tracker.IsInCooldown("key1") {
		t.Error("empty key should not be in cooldown")
	}
}

func TestTracker_RecordAndIsInCooldown(t *testing.T) {
	tracker := NewTracker(1, 70, false) // 1 minute cooldown
	key := "user@x.com:device1"

	tracker.Record(key)
	if !tracker.IsInCooldown(key) {
		t.Error("key should be in cooldown immediately after Record")
	}
}

func TestTracker_Reset(t *testing.T) {
	tracker := NewTracker(60, 70, false)
	key := "user@x.com:device1"

	tracker.Record(key)
	if !tracker.IsInCooldown(key) {
		t.Error("expected in cooldown after Record")
	}
	tracker.Reset(key)
	if tracker.IsInCooldown(key) {
		t.Error("should not be in cooldown after Reset")
	}
}

func TestTracker_ShouldReset(t *testing.T) {
	tracker := NewTracker(60, 70, false)

	if !tracker.ShouldReset(50) {
		t.Error("50% should be below 70% reset threshold")
	}
	if !tracker.ShouldReset(69) {
		t.Error("69% should be below 70% reset threshold")
	}
	if tracker.ShouldReset(70) {
		t.Error("70% should not trigger reset")
	}
	if tracker.ShouldReset(80) {
		t.Error("80% should not trigger reset")
	}
}

func TestTracker_CooldownExpiry(t *testing.T) {
	tracker := NewTracker(0, 70, false) // 0 minute cooldown - effectively instant expiry
	key := "user@x.com:device1"

	tracker.Record(key)
	// With 0 cooldown, time.Since(lastTime) will exceed 0 immediately
	time.Sleep(10 * time.Millisecond)
	if tracker.IsInCooldown(key) {
		t.Error("with 0 cooldown, should not be in cooldown after brief wait")
	}
}
