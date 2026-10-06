package idle

import "time"

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

// Poll runs one decision cycle and returns (fired, roastMsg).
//
// fired=false means the idle threshold/repeat check didn't trigger — caller
// does nothing. fired=true with a non-empty roastMsg means a media player was
// detected and a roast was composed — use roastMsg verbatim. fired=true with
// an empty roastMsg means the threshold fired but no media player was active —
// the caller should render its own nudge text from its template pool.
//
// project and dormantFor are passed into the roast context so the roast can
// mention which project is sitting dormant while the user watches something.
// Pass "" and 0 when the caller has no project context.
//
// Media-player state never suppresses the nudge — Watcher.Decide's
// idle/threshold/repeat logic is untouched. A MediaPlayerReader error is
// treated as "nothing playing" so a flaky D-Bus call never silences a nudge.
func (e *NudgeEngine) Poll(idleTime time.Duration, lastNudged time.Time, now time.Time, project string, dormantFor time.Duration) (fired bool, roastMsg string) {
	return e.PollContext(idleTime, lastNudged, now, NudgeContext{
		Project:    project,
		DormantFor: dormantFor,
	})
}

// PollContext runs one decision cycle with full NudgeContext (including uncommitted status).
func (e *NudgeEngine) PollContext(idleTime time.Duration, lastNudged time.Time, now time.Time, ctx NudgeContext) (fired bool, roastMsg string) {
	if e == nil || e.watcher == nil {
		return false, ""
	}
	if !e.watcher.Decide(idleTime, lastNudged, now) {
		return false, ""
	}

	if e.media != nil {
		if playing, player, err := e.media.NowPlaying(); err == nil && playing {
			if e.roaster != nil {
				msg := e.roaster.Compose(ctx, player)
				return true, msg
			}
		}
	}

	// Fired, but no roast — caller picks the template.
	return true, ""
}
