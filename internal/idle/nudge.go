package idle

import (
	"fmt"
	"time"
)

// NudgeEngine is the single thing Sentinel's poll loop should call each
// tick. It wraps Watcher (idle edge-detection, unchanged from decide.go)
// with MediaPlayerReader (is a video/audio player active right now) and
// Roaster (what to say about it), so the poll loop doesn't need to know
// about any of those three separately.
type NudgeEngine struct {
	watcher *Watcher
	media   MediaPlayerReader
	roaster *Roaster
}

// NewNudgeEngine wires the three components together.
func NewNudgeEngine(watcher *Watcher, media MediaPlayerReader, roaster *Roaster) *NudgeEngine {
	return &NudgeEngine{watcher: watcher, media: media, roaster: roaster}
}

// Poll runs one decision cycle and returns (fire, message). message is
// empty when fire is false. project/dormantFor describe the dormant
// project Sentinel wants to nudge about (from ProjectManager).
//
// Media-player state never suppresses the nudge — Watcher.Decide's
// idle/threshold/repeat logic is untouched, so this is purely additive.
// If Decide says fire and a player is actively playing (not just open —
// paused doesn't count), the roast calls it out by name; otherwise a
// plain nudge fires. A MediaPlayerReader error is treated the same as
// "nothing playing" rather than blocking the nudge — a flaky D-Bus call
// shouldn't be the reason Sentinel goes quiet.
func (e *NudgeEngine) Poll(idleTime time.Duration, lastNudged time.Time, now time.Time, project string, dormantFor time.Duration) (bool, string) {
	if e == nil || e.watcher == nil {
		return false, ""
	}
	if !e.watcher.Decide(idleTime, lastNudged, now) {
		return false, ""
	}

	if e.media != nil {
		if playing, player, err := e.media.NowPlaying(); err == nil && playing {
			if e.roaster != nil {
				msg := e.roaster.Compose(NudgeContext{Project: project, DormantFor: dormantFor}, player)
				return true, msg
			}
		}
	}

	if project != "" {
		return true, fmt.Sprintf("%d minutes idle. %s is still waiting.", int(idleTime.Minutes()), project)
	}
	return true, fmt.Sprintf("%d minutes idle.", int(idleTime.Minutes()))
}
