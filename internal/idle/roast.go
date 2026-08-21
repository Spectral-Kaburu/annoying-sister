package idle

import (
	"fmt"
	"math/rand"
	"time"
)

// NudgeContext carries the info a roast needs: which project has gone
// dormant (from Sentinel/ProjectManager) and how long it's been idle.
type NudgeContext struct {
	Project    string
	DormantFor time.Duration
}

// roastFunc renders one line of a roast given the nudge context and the
// human-friendly player name from MediaPlayerReader.NowPlaying.
type roastFunc func(ctx NudgeContext, player string) string

// roastTemplates: sharp, specific to the procrastination (player X open,
// project Y untouched), not attacks on the person. Add/replace freely —
// nothing else in the package depends on their exact wording, only on
// Roaster.Compose returning a non-empty string.
var roastTemplates = []roastFunc{
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "your project"
		}
		return fmt.Sprintf(
			"%d minutes idle, %s open, %s untouched. Bold strategy, considering %s isn't going to finish itself.",
			int(ctx.DormantFor.Minutes()), player, proj, proj,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "your project"
		}
		return fmt.Sprintf(
			"You didn't get distracted — you *chose* %s over %s. At least own it while you go fix it.",
			player, proj,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "your project"
		}
		return fmt.Sprintf(
			"%s: paused. %s: playing. Somewhere your future self is filing this under 'foreshadowing.'",
			proj, player,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "the project"
		}
		return fmt.Sprintf(
			"Noted for the %s post-mortem: 'lost to %s, %d minutes in, no survivors.'",
			proj, player, int(ctx.DormantFor.Minutes()),
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "Your project"
		}
		return fmt.Sprintf(
			"%s has been sitting dormant long enough to grow its own dust layer, and you picked *now* to catch up on %s. Incredible timing.",
			proj, player,
		)
	},
}

// Roaster picks a roast template at random. Seed it explicitly (rather
// than using the global rand source) so tests can construct a
// deterministic Roaster and assert on exact output.
type Roaster struct {
	rng *rand.Rand
}

// NewRoaster builds a Roaster seeded with seed. Pass time.Now().UnixNano()
// in production for variety; pass a fixed constant in tests.
func NewRoaster(seed int64) *Roaster {
	return &Roaster{rng: rand.New(rand.NewSource(seed))}
}

// Compose picks a random roast template and renders it against ctx and
// the currently-playing player's name.
func (r *Roaster) Compose(ctx NudgeContext, player string) string {
	if r == nil || r.rng == nil {
		r = NewRoaster(time.Now().UnixNano())
	}
	tmpl := roastTemplates[r.rng.Intn(len(roastTemplates))]
	return tmpl(ctx, player)
}
