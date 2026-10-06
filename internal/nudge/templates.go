package nudge

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/Spectral-Kaburu/annoying-sister/internal/paths"
	"github.com/Spectral-Kaburu/annoying-sister/internal/prompts"
)

// Slots holds every placeholder value a template pool might reference.
// Not every field is used by every category.
type Slots struct {
	ProjectName      string // never a file path — see Content Restrictions in the spec
	IdleMinutes      int
	DormantDays      int
	HasUncommitted   bool
	UncommittedCount int
}

var defaultStore *prompts.Store

func getStore() *prompts.Store {
	if defaultStore != nil {
		return defaultStore
	}
	pDir, err := paths.PromptsDir()
	if err == nil {
		defaultStore = prompts.NewStore(pDir)
		return defaultStore
	}
	return prompts.NewStore("")
}

// SetStore sets a custom prompt store for rendering (useful in tests or custom setups).
func SetStore(s *prompts.Store) {
	defaultStore = s
}

func pick(rng *rand.Rand, pool []string) string {
	if len(pool) == 0 {
		return ""
	}
	return pool[rng.Intn(len(pool))]
}

func fill(template string, slots Slots) string {
	proj := slots.ProjectName
	if proj != "" && !strings.HasPrefix(strings.ToLower(proj), "project ") {
		proj = "project " + proj
	}
	r := strings.NewReplacer(
		"{ProjectName}", proj,
		"{IdleMinutes}", fmt.Sprintf("%d", slots.IdleMinutes),
		"{DormantDays}", fmt.Sprintf("%d", slots.DormantDays),
		"{UncommittedCount}", fmt.Sprintf("%d", slots.UncommittedCount),
	)
	res := r.Replace(template)
	if strings.HasPrefix(res, "project ") {
		res = "Project " + strings.TrimPrefix(res, "project ")
	}
	return res
}

// RenderIdle picks a random idle-pool template and fills it with slots.IdleMinutes.
func RenderIdle(rng *rand.Rand, slots Slots) string {
	pool := getStore().Idle()
	if len(pool) == 0 {
		return fmt.Sprintf("Still there? It's been %d minutes of total silence.", slots.IdleMinutes)
	}
	return fill(pick(rng, pool), slots)
}

// RenderDormant picks a random dormant-project-pool template. If
// slots.HasUncommitted is true, it selects from the uncommitted pool.
func RenderDormant(rng *rand.Rand, slots Slots) string {
	var pool []string
	if slots.HasUncommitted {
		pool = getStore().DormantUncommitted()
	} else {
		pool = getStore().Dormant()
	}
	if len(pool) == 0 {
		if slots.HasUncommitted {
			return fill("{ProjectName} has uncommitted changes sitting untouched for {DormantDays} days.", slots)
		}
		return fill("{ProjectName} hasn't seen a single commit or edit in {DormantDays} days.", slots)
	}
	return fill(pick(rng, pool), slots)
}

// RenderOnStart picks a random on-start template.
func RenderOnStart(rng *rand.Rand, hasDormant bool, slots Slots) string {
	var pool []string
	if !hasDormant {
		pool = getStore().StartupGeneric()
	} else if slots.HasUncommitted {
		pool = getStore().StartupUncommitted()
	} else {
		pool = getStore().StartupDormant()
	}
	if len(pool) == 0 {
		return "Nudge daemon online and watching your back."
	}
	return fill(pick(rng, pool), slots)
}
