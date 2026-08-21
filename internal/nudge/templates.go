package nudge

import (
	"fmt"
	"math/rand"
	"strings"
)

// Slots holds every placeholder value a template pool might reference.
// Not every field is used by every category.
type Slots struct {
	ProjectName string // never a file path — see Content Restrictions in the spec
	IdleMinutes int
	DormantDays int
}

// Placeholder template pools. Per the spec these are explicitly marked
// "rewrite before ship" — what matters structurally right now is that each
// pool has enough entries that repeated firings (especially idle, which
// can repeat many times in one stretch) don't read as identical.
var (
	idlePool = []string{
		"Still there? It's been {IdleMinutes} minutes.",
		"{IdleMinutes} minutes of silence. Riveting.",
		"I'll just assume you're thinking very deeply for {IdleMinutes} minutes.",
		"{IdleMinutes} minutes and counting. The cursor blinks alone.",
		"Taking a break, or is this permanent? {IdleMinutes} minutes so far.",
	}

	dormantPool = []string{
		"{ProjectName} hasn't heard from you in {DormantDays} days. It's starting to worry.",
		"{DormantDays} days since you touched {ProjectName}. Just saying.",
		"Remember {ProjectName}? It remembers you.",
		"{ProjectName} has been sitting untouched for {DormantDays} days now.",
		"It's been {DormantDays} days. {ProjectName} isn't going to finish itself.",
	}

	// Used when at least one project is dormant at startup.
	onStartWithDormantPool = []string{
		"I'm up. {ProjectName} is still waiting, by the way.",
		"Starting fresh. Nothing's changed with {ProjectName} — {DormantDays} days and counting.",
		"Back online. {ProjectName} has been quiet for {DormantDays} days.",
	}

	// Used when no project is currently dormant at startup.
	onStartGenericPool = []string{
		"I'm up and watching.",
		"Nudge is running. Let's see how long that lasts.",
		"Started up. Everything looks current for now.",
	}
)

func pick(rng *rand.Rand, pool []string) string {
	return pool[rng.Intn(len(pool))]
}

func fill(template string, slots Slots) string {
	r := strings.NewReplacer(
		"{ProjectName}", slots.ProjectName,
		"{IdleMinutes}", fmt.Sprintf("%d", slots.IdleMinutes),
		"{DormantDays}", fmt.Sprintf("%d", slots.DormantDays),
	)
	return r.Replace(template)
}

// RenderIdle picks a random idle-pool template and fills it with
// slots.IdleMinutes.
func RenderIdle(rng *rand.Rand, slots Slots) string {
	return fill(pick(rng, idlePool), slots)
}

// RenderDormant picks a random dormant-project-pool template and fills it
// with slots.ProjectName and slots.DormantDays.
func RenderDormant(rng *rand.Rand, slots Slots) string {
	return fill(pick(rng, dormantPool), slots)
}

// RenderOnStart picks a random on-start template. If hasDormant is false,
// the generic pool is used regardless of what's in slots, since there is
// no dormant project to reference.
func RenderOnStart(rng *rand.Rand, hasDormant bool, slots Slots) string {
	if !hasDormant {
		return fill(pick(rng, onStartGenericPool), slots)
	}
	return fill(pick(rng, onStartWithDormantPool), slots)
}
