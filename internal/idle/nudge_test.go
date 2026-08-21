package idle

import (
	"strings"
	"testing"
	"time"
)

// fakeMediaPlayerReader lets tests control NowPlaying's return value
// without a real D-Bus session, mirroring how decide_test.go drives
// Watcher directly with plain arguments instead of a real Reader.
type fakeMediaPlayerReader struct {
	playing bool
	player  string
	err     error
}

func (f fakeMediaPlayerReader) NowPlaying() (bool, string, error) {
	return f.playing, f.player, f.err
}

func TestNudgeEngine_BelowThreshold_NeverFires(t *testing.T) {
	watcher := NewWatcher(20*time.Minute, 20*time.Minute)
	media := fakeMediaPlayerReader{playing: true, player: "Vlc"}
	engine := NewNudgeEngine(watcher, media, NewRoaster(1))

	fire, msg := engine.Poll(5*time.Minute, time.Time{}, time.Now(), "SpectreTTS", 5*time.Minute)

	if fire {
		t.Errorf("expected no fire below threshold, got fire=true msg=%q", msg)
	}
	if msg != "" {
		t.Errorf("expected empty message when not firing, got %q", msg)
	}
}

func TestNudgeEngine_FiresPlain_WhenNoPlayerActive(t *testing.T) {
	watcher := NewWatcher(20*time.Minute, 20*time.Minute)
	media := fakeMediaPlayerReader{playing: false}
	engine := NewNudgeEngine(watcher, media, NewRoaster(1))

	now := time.Now()
	fire, msg := engine.Poll(20*time.Minute, time.Time{}, now, "SpectreTTS", 20*time.Minute)

	if !fire {
		t.Fatal("expected fire at threshold")
	}
	if msg == "" {
		t.Fatal("expected non-empty plain nudge message")
	}
	if strings.Contains(msg, "Vlc") {
		t.Errorf("plain nudge should not reference a player, got %q", msg)
	}
}

func TestNudgeEngine_FiresRoast_WhenPlayerActive(t *testing.T) {
	watcher := NewWatcher(20*time.Minute, 20*time.Minute)
	media := fakeMediaPlayerReader{playing: true, player: "Vlc"}
	engine := NewNudgeEngine(watcher, media, NewRoaster(1))

	now := time.Now()
	fire, msg := engine.Poll(20*time.Minute, time.Time{}, now, "SpectreTTS", 20*time.Minute)

	if !fire {
		t.Fatal("expected fire at threshold")
	}
	if !strings.Contains(msg, "Vlc") {
		t.Errorf("expected roast to mention active player, got %q", msg)
	}
	if !strings.Contains(msg, "SpectreTTS") {
		t.Errorf("expected roast to mention project, got %q", msg)
	}
}

func TestNudgeEngine_MediaPlayerError_FallsBackToPlainNudge(t *testing.T) {
	// A flaky D-Bus call on the media-player side must not swallow
	// the nudge entirely — it should just degrade to the plain
	// message, same as "nothing playing".
	watcher := NewWatcher(20*time.Minute, 20*time.Minute)
	media := fakeMediaPlayerReader{playing: true, player: "Vlc", err: errBusUnavailable}
	engine := NewNudgeEngine(watcher, media, NewRoaster(1))

	now := time.Now()
	fire, msg := engine.Poll(20*time.Minute, time.Time{}, now, "SpectreTTS", 20*time.Minute)

	if !fire {
		t.Fatal("expected fire at threshold despite media-player error")
	}
	if strings.Contains(msg, "Vlc") {
		t.Errorf("expected plain fallback nudge on media-player error, got %q", msg)
	}
}

var errBusUnavailable = &fakeDBusError{"bus unavailable"}

type fakeDBusError struct{ msg string }

func (e *fakeDBusError) Error() string { return e.msg }
