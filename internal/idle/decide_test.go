package idle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcher_BelowThreshold_NeverFires(t *testing.T) {
	w := NewWatcher(20*time.Minute, 20*time.Minute)
	now := time.Now()
	assert.False(t, w.Decide(5*time.Minute, time.Time{}, now))
}

func TestWatcher_CrossingThreshold_FiresOnce(t *testing.T) {
	w := NewWatcher(20*time.Minute, 20*time.Minute)
	now := time.Now()

	// Below threshold: no fire.
	assert.False(t, w.Decide(19*time.Minute, time.Time{}, now))
	// Crosses threshold: fires (the "first" nudge of the stretch).
	assert.True(t, w.Decide(20*time.Minute, time.Time{}, now))
}

func TestWatcher_StaysAboveThreshold_DoesNotRefireBeforeRepeatInterval(t *testing.T) {
	w := NewWatcher(20*time.Minute, 20*time.Minute)
	now := time.Now()

	// First poll crosses the threshold and fires the "first" nudge.
	require.True(t, w.Decide(20*time.Minute, time.Time{}, now))
	lastNudged := now

	// Still idle 5 minutes later, repeat interval (20m) hasn't elapsed.
	fired := w.Decide(25*time.Minute, lastNudged, now.Add(5*time.Minute))
	assert.False(t, fired)
}

func TestWatcher_StaysAboveThreshold_RefiresAfterRepeatInterval(t *testing.T) {
	w := NewWatcher(20*time.Minute, 20*time.Minute)
	now := time.Now()

	assert.True(t, w.Decide(20*time.Minute, time.Time{}, now)) // first nudge
	lastNudged := now

	// 20+ minutes later, still idle: repeat fires.
	fired := w.Decide(40*time.Minute, lastNudged, now.Add(20*time.Minute))
	assert.True(t, fired)
}

func TestWatcher_DropBelowThreshold_ResetsEdgeForNextStretch(t *testing.T) {
	w := NewWatcher(20*time.Minute, 20*time.Minute)
	now := time.Now()

	assert.True(t, w.Decide(20*time.Minute, time.Time{}, now)) // first nudge of stretch 1
	// User becomes active again.
	assert.False(t, w.Decide(0, now, now.Add(1*time.Minute)))

	// A brand-new idle stretch immediately crosses threshold again: this
	// must be treated as a fresh "first" nudge, not a stale repeat check
	// (even though very little time has passed since the last nudge).
	fired := w.Decide(20*time.Minute, now, now.Add(2*time.Minute))
	assert.True(t, fired, "a new idle stretch must fire immediately, not wait out the repeat interval")
}

func TestWatcher_RepeatWhileIdle_CheckedAgainstLastNudged_NotASeparateEdge(t *testing.T) {
	// Regression guard: repeats must be judged purely against
	// last_nudged + repeat_interval, not against some other internal
	// "still idle" edge timer.
	w := NewWatcher(10*time.Minute, 5*time.Minute)
	now := time.Now()

	assert.True(t, w.Decide(10*time.Minute, time.Time{}, now))
	lastNudged := now

	// Multiple polls while still idle, before repeat interval elapses:
	// none should fire.
	assert.False(t, w.Decide(11*time.Minute, lastNudged, now.Add(1*time.Minute)))
	assert.False(t, w.Decide(12*time.Minute, lastNudged, now.Add(2*time.Minute)))
	assert.False(t, w.Decide(14*time.Minute, lastNudged, now.Add(4*time.Minute)))
	// Exactly at the repeat interval: fires.
	assert.True(t, w.Decide(15*time.Minute, lastNudged, now.Add(5*time.Minute)))
}
