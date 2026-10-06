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
		"Still there? It's been {IdleMinutes} minutes of total silence.",
		"{IdleMinutes} minutes without a single keystroke. The blinking cursor is weeping.",
		"I'll just assume you're having an intense philosophical breakthrough for {IdleMinutes} minutes.",
		"Taking a quick break or quietly retiring? You've been idle for {IdleMinutes} minutes.",
		"{IdleMinutes} minutes idle. Your keyboard is filing a missing persons report.",
		"Hello? Earth to developer. {IdleMinutes} minutes have vanished into the void.",
		"Just checking in. It's been {IdleMinutes} minutes of absolute stillness.",
		"{IdleMinutes} minutes of staring at the screen. Either you're in deep thought, or you fell asleep.",
	}

	dormantPool = []string{
		"{ProjectName} hasn't seen a single commit or edit in {DormantDays} days. It's getting lonely.",
		"{DormantDays} days since you last touched {ProjectName}. Remember when you said you'd finish it this week?",
		"{ProjectName} has been collecting digital dust for {DormantDays} days now.",
		"It's been {DormantDays} days. {ProjectName} isn't going to build, test, or ship itself.",
		"Quick reminder: {ProjectName} is still sitting there untouched after {DormantDays} days.",
		"{ProjectName} is slowly turning into an archaeological artifact. {DormantDays} days untouched.",
		"Hey, {ProjectName} is quietly judging your life choices after {DormantDays} days of silence.",
	}

	// Used when at least one project is dormant at startup.
	onStartWithDormantPool = []string{
		"I'm awake and watching. By the way, {ProjectName} has been waiting on you for {DormantDays} days.",
		"System online. Just so you know, {ProjectName} hasn't moved in {DormantDays} days.",
		"Daemon started. Let's see if we can finally give {ProjectName} some attention after {DormantDays} days.",
		"Online and monitoring. Friendly heads up: {ProjectName} is {DormantDays} days dormant.",
	}

	// Used when no project is currently dormant at startup.
	onStartGenericPool = []string{
		"Nudge daemon online and watching your back.",
		"Systems active. All projects look fresh. Let's keep it that way.",
		"Daemon started. Ready to keep you honest.",
		"I'm awake. Don't slack off today.",
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
