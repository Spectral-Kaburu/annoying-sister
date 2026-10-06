package nudge

import (
	"fmt"
	"math/rand"
	"strings"
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

// Placeholder template pools.
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

	dormantWithUncommittedPool = []string{
		"{ProjectName} has uncommitted changes sitting untouched for {DormantDays} days. Don't lose your train of thought.",
		"You left {UncommittedCount} uncommitted files in {ProjectName} {DormantDays} days ago. It's so close to being done.",
		"Reminder: {ProjectName} still has uncommitted work waiting on you from {DormantDays} days ago.",
		"{ProjectName} has uncommitted edits waiting for a commit. Finish what you started {DormantDays} days ago.",
		"{DormantDays} days since you made edits in {ProjectName} without committing. You know you'll forget what they do.",
	}

	// Used when at least one project is dormant at startup with uncommitted changes.
	onStartWithUncommittedPool = []string{
		"I'm awake and watching. Heads up: you left uncommitted changes in {ProjectName} from {DormantDays} days ago.",
		"System online. {ProjectName} has {UncommittedCount} uncommitted files waiting to be wrapped up.",
		"Daemon started. You have uncommitted work in {ProjectName}. Let's finish it today.",
	}

	// Used when at least one project is dormant at startup without uncommitted changes.
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

// RenderIdle picks a random idle-pool template and fills it with
// slots.IdleMinutes.
func RenderIdle(rng *rand.Rand, slots Slots) string {
	return fill(pick(rng, idlePool), slots)
}

// RenderDormant picks a random dormant-project-pool template. If
// slots.HasUncommitted is true, it selects from the uncommitted pool.
func RenderDormant(rng *rand.Rand, slots Slots) string {
	if slots.HasUncommitted {
		return fill(pick(rng, dormantWithUncommittedPool), slots)
	}
	return fill(pick(rng, dormantPool), slots)
}

// RenderOnStart picks a random on-start template.
func RenderOnStart(rng *rand.Rand, hasDormant bool, slots Slots) string {
	if !hasDormant {
		return fill(pick(rng, onStartGenericPool), slots)
	}
	if slots.HasUncommitted {
		return fill(pick(rng, onStartWithUncommittedPool), slots)
	}
	return fill(pick(rng, onStartWithDormantPool), slots)
}
