package nudge

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderIdle_FillsSlot(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	text := RenderIdle(rng, Slots{IdleMinutes: 42})
	assert.Contains(t, text, "42")
	assert.NotContains(t, text, "{IdleMinutes}")
}

func TestRenderDormant_FillsSlots(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	text := RenderDormant(rng, Slots{ProjectName: "VidKing", DormantDays: 9})
	assert.Contains(t, text, "VidKing")
	assert.Contains(t, text, "9")
	assert.NotContains(t, text, "{ProjectName}")
	assert.NotContains(t, text, "{DormantDays}")
}

func TestRenderOnStart_NoDormant_UsesGenericPool(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 20; i++ {
		text := RenderOnStart(rng, false, Slots{})
		assert.NotContains(t, text, "{ProjectName}")
		assert.NotContains(t, text, "{DormantDays}")
	}
}

func TestRenderOnStart_WithDormant_MentionsProject(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	text := RenderOnStart(rng, true, Slots{ProjectName: "VidKing", DormantDays: 30})
	assert.Contains(t, text, "VidKing")
}

func TestRenderDormant_WithUncommitted_MentionsUncommitted(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	text := RenderDormant(rng, Slots{
		ProjectName:      "VidKing",
		DormantDays:      5,
		HasUncommitted:   true,
		UncommittedCount: 3,
	})
	assert.Contains(t, text, "VidKing")
	assert.Contains(t, text, "uncommitted")
	assert.NotContains(t, text, "{ProjectName}")
	assert.NotContains(t, text, "{UncommittedCount}")
}

func TestRenderOnStart_WithUncommitted_MentionsUncommitted(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	text := RenderOnStart(rng, true, Slots{
		ProjectName:      "VidKing",
		DormantDays:      5,
		HasUncommitted:   true,
		UncommittedCount: 4,
	})
	assert.Contains(t, text, "VidKing")
	assert.Contains(t, text, "uncommitted")
}

// Regression guard for the spec's explicit requirement: repeated idle
// firings within one stretch must not consistently read as the same line.
// A single fixed template would make every idle nudge identical; this
// asserts the pool actually has enough variety to avoid that in practice.
func TestIdlePool_HasEnoughVarietyForRepeatedFirings(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := getStore().Idle()
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		seen[pick(rng, pool)] = true
	}
	require.Greater(t, len(pool), 1, "idle pool must have more than one template")
	require.Greater(t, len(seen), 1, "200 draws from the idle pool should surface more than one distinct template")
}
