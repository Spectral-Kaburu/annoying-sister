package idle

import (
	"strings"
	"testing"
	"time"
)

func TestRoaster_Compose_IsDeterministicForFixedSeed(t *testing.T) {
	ctx := NudgeContext{Project: "SpectreTTS", DormantFor: 45 * time.Minute}

	r1 := NewRoaster(42)
	r2 := NewRoaster(42)

	got1 := r1.Compose(ctx, "Vlc")
	got2 := r2.Compose(ctx, "Vlc")

	if got1 != got2 {
		t.Errorf("same seed produced different output:\n%q\n%q", got1, got2)
	}
}

func TestRoaster_Compose_EveryTemplateMentionsProjectAndPlayer(t *testing.T) {
	// Drive Compose enough times with a fixed seed to hit every
	// template at least once (5 templates; 50 draws makes an
	// all-miss vanishingly unlikely) and check each one actually
	// contains the project/player it was given rather than a stale
	// hardcoded string like "VLC" from an earlier draft.
	ctx := NudgeContext{Project: "Linus", DormantFor: 20 * time.Minute}
	r := NewRoaster(7)

	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		msg := r.Compose(ctx, "Mpv")
		seen[msg] = true

		if !strings.Contains(msg, "Linus") {
			t.Errorf("roast missing project name: %q", msg)
		}
		if !strings.Contains(msg, "Mpv") {
			t.Errorf("roast missing player name: %q", msg)
		}
	}

	if len(seen) < 2 {
		t.Errorf("expected variety across draws, got %d distinct messages", len(seen))
	}
}

func TestRoaster_Compose_WithUncommittedChanges(t *testing.T) {
	ctx := NudgeContext{
		Project:          "AnnoyingSister",
		DormantFor:       3 * 24 * time.Hour,
		HasUncommitted:   true,
		UncommittedCount: 5,
	}
	r := NewRoaster(123)
	msg := r.Compose(ctx, "Spotify")
	if !strings.Contains(msg, "AnnoyingSister") {
		t.Errorf("expected roast to mention project, got %q", msg)
	}
	if !strings.Contains(msg, "Spotify") {
		t.Errorf("expected roast to mention player, got %q", msg)
	}
}
