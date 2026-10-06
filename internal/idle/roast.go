package idle

import (
	"fmt"
	"math/rand"
	"time"
)

// NudgeContext carries the info a roast needs: which project has gone
// dormant (from Sentinel/ProjectManager) and how long it's been idle.
type NudgeContext struct {
	Project          string
	DormantFor       time.Duration
	HasUncommitted   bool
	UncommittedCount int
}

// roastFunc renders one line of a roast given the nudge context and the
// human-friendly player name from MediaPlayerReader.NowPlaying.
type roastFunc func(ctx NudgeContext, player string) string

// roastTemplates: sharp, specific to the procrastination (player X open,
// project Y untouched), not attacks on the person.
var roastTemplates = []roastFunc{
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "your project"
		}
		if ctx.HasUncommitted {
			return fmt.Sprintf(
				"%s is playing while %s has %d uncommitted files waiting. Bold strategy, considering code doesn't commit itself.",
				player, proj, ctx.UncommittedCount,
			)
		}
		if ctx.DormantFor > 0 {
			return fmt.Sprintf(
				"%d days since you touched %s, and now you're bingeing on %s. Bold strategy, let's see how it plays out.",
				int(ctx.DormantFor.Hours()/24), proj, player,
			)
		}
		return fmt.Sprintf(
			"%s is playing while %s sits completely untouched. Bold strategy, considering code doesn't write itself.",
			player, proj,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "your project"
		}
		if ctx.HasUncommitted {
			return fmt.Sprintf(
				"You didn't just walk away from %s — you left uncommitted work sitting there while you watch %s. At least own it.",
				proj, player,
			)
		}
		return fmt.Sprintf(
			"You didn't get distracted — you made a conscious decision to choose %s over %s. At least own it.",
			player, proj,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "your project"
		}
		if ctx.HasUncommitted {
			return fmt.Sprintf(
				"%s: uncommitted edits pending. %s: currently playing. Priorities, right?",
				proj, player,
			)
		}
		return fmt.Sprintf(
			"%s: paused. %s: playing. Somewhere your future self is filing this under 'procrastination evidence.'",
			proj, player,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "the project"
		}
		if ctx.HasUncommitted {
			return fmt.Sprintf(
				"Noted for the %s post-mortem: 'abandoned mid-edit with uncommitted changes for %s, no survivors.'",
				proj, player,
			)
		}
		if ctx.DormantFor > 0 {
			return fmt.Sprintf(
				"Noted for the %s post-mortem: 'lost to %s after %d days of inactivity, no survivors.'",
				proj, player, int(ctx.DormantFor.Hours()/24),
			)
		}
		return fmt.Sprintf(
			"Noted for the %s post-mortem: 'lost to %s, zero progress made.'",
			proj, player,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "Your project"
		}
		if ctx.HasUncommitted {
			return fmt.Sprintf(
				"%s is waiting for you to commit your unfinished changes, and you picked right now to enjoy %s. Incredible timing.",
				proj, player,
			)
		}
		return fmt.Sprintf(
			"%s has been sitting dormant long enough to develop its own ecosystem, and you picked right now to enjoy %s. Incredible timing.",
			proj, player,
		)
	},
	func(ctx NudgeContext, player string) string {
		proj := ctx.Project
		if proj == "" {
			proj = "your project"
		}
		return fmt.Sprintf(
			"Ah yes, %s is definitely the prerequisite dependency required to ship %s. Keep telling yourself that.",
			player, proj,
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
