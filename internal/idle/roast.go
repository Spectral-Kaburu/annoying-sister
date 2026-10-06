package idle

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Spectral-Kaburu/annoying-sister/internal/paths"
	"github.com/Spectral-Kaburu/annoying-sister/internal/prompts"
)

// NudgeContext carries the info a roast needs: which project has gone
// dormant (from Sentinel/ProjectManager) and how long it's been idle.
type NudgeContext struct {
	Project          string
	DormantFor       time.Duration
	HasUncommitted   bool
	UncommittedCount int
}

func formatProjectName(proj string) string {
	if proj == "" {
		return "your project"
	}
	if strings.HasPrefix(strings.ToLower(proj), "project ") || strings.HasPrefix(strings.ToLower(proj), "your project") {
		return proj
	}
	return "project " + proj
}

func fillRoast(template string, ctx NudgeContext, player string) string {
	proj := formatProjectName(ctx.Project)
	r := strings.NewReplacer(
		"{Player}", player,
		"{ProjectName}", proj,
		"{DormantDays}", fmt.Sprintf("%d", int(ctx.DormantFor.Hours()/24)),
		"{UncommittedCount}", fmt.Sprintf("%d", ctx.UncommittedCount),
	)
	res := r.Replace(template)
	if strings.HasPrefix(res, "project ") {
		res = "Project " + strings.TrimPrefix(res, "project ")
	}
	return res
}

// Roaster picks a roast template at random. Seed it explicitly (rather
// than using the global rand source) so tests can construct a
// deterministic Roaster and assert on exact output.
type Roaster struct {
	rng   *rand.Rand
	store *prompts.Store
}

// NewRoaster builds a Roaster seeded with seed. Pass time.Now().UnixNano()
// in production for variety; pass a fixed constant in tests.
func NewRoaster(seed int64) *Roaster {
	var store *prompts.Store
	if pDir, err := paths.PromptsDir(); err == nil {
		store = prompts.NewStore(pDir)
	} else {
		store = prompts.NewStore("")
	}
	return &Roaster{
		rng:   rand.New(rand.NewSource(seed)),
		store: store,
	}
}

// NewRoasterWithStore builds a Roaster with an explicit prompt store.
func NewRoasterWithStore(seed int64, store *prompts.Store) *Roaster {
	return &Roaster{
		rng:   rand.New(rand.NewSource(seed)),
		store: store,
	}
}

// Compose picks a random roast template and renders it against ctx and
// the currently-playing player's name.
func (r *Roaster) Compose(ctx NudgeContext, player string) string {
	if r == nil || r.rng == nil {
		r = NewRoaster(time.Now().UnixNano())
	}
	store := r.store
	if store == nil {
		if pDir, err := paths.PromptsDir(); err == nil {
			store = prompts.NewStore(pDir)
		} else {
			store = prompts.NewStore("")
		}
	}

	var pool []string
	if ctx.HasUncommitted {
		pool = store.RoastsUncommitted()
	} else if ctx.DormantFor >= 24*time.Hour {
		pool = store.RoastsDormant()
	} else {
		pool = store.RoastsIdle()
	}

	if len(pool) == 0 {
		pool = store.RoastsIdle()
	}
	if len(pool) == 0 {
		pool = []string{"{Player} is playing while {ProjectName} sits completely untouched."}
	}

	tmpl := pool[r.rng.Intn(len(pool))]
	return fillRoast(tmpl, ctx, player)
}
