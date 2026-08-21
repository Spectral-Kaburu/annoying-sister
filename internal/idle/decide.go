package idle

import "time"

// Watcher holds the small piece of state needed to implement the idle
// nudge's "repeat-while-condition-holds, reset on drop-below-threshold"
// behavior described in the spec. It takes plain idle-time/timestamp
// arguments and returns a decision — it never touches D-Bus or the
// filesystem itself, so it's unit-testable in isolation.
type Watcher struct {
	Threshold      time.Duration
	RepeatInterval time.Duration

	// wasAboveThreshold tracks whether the previous poll observed idle
	// time at/above Threshold. It is the edge-detection state: false ->
	// true is what fires the "first" nudge of a stretch.
	wasAboveThreshold bool
}

// NewWatcher constructs a Watcher with the given threshold/repeat
// durations. Edge-detection state starts as "not idle".
func NewWatcher(threshold, repeatInterval time.Duration) *Watcher {
	return &Watcher{Threshold: threshold, RepeatInterval: repeatInterval}
}

// Decide reports whether an idle nudge should fire on this poll, given the
// current idleTime, the timestamp of the last idle nudge (lastNudged, zero
// value if none has ever fired), and the current time now.
//
// Behavior, matching the spec exactly:
//   - idleTime below Threshold: reset edge-detection state (so the next
//     stretch gets a fresh "first" nudge) and never fire.
//   - idleTime at/above Threshold, previous poll was below: this is the
//     transition edge — fire (the "first" nudge of the stretch).
//   - idleTime at/above Threshold, previous poll was also at/above: a
//     repeat-while-idle check against lastNudged, not a separate edge —
//     fire only if RepeatInterval has elapsed since lastNudged.
func (w *Watcher) Decide(idleTime time.Duration, lastNudged time.Time, now time.Time) bool {
	if idleTime < w.Threshold {
		w.wasAboveThreshold = false
		return false
	}

	if !w.wasAboveThreshold {
		w.wasAboveThreshold = true
		return true
	}

	return now.Sub(lastNudged) >= w.RepeatInterval
}
