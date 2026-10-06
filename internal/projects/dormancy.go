package projects

import "time"

// IsDormant reports whether p has gone longer than threshold since
// LastActive, relative to now.
func IsDormant(p Project, threshold time.Duration, now time.Time) bool {
	return now.Sub(p.LastActive) > threshold
}

// MostDormant returns the name and entry of the project with the oldest
// LastActive timestamp in reg, and true if reg is non-empty.
func MostDormant(reg Registry) (name string, project Project, ok bool) {
	first := true
	for k, p := range reg {
		if first || p.LastActive.Before(project.LastActive) {
			name, project = k, p
			first = false
		}
	}
	return name, project, !first
}

// MostUrgent selects the highest priority project to nudge based on dormancy and uncommitted changes.
//
// Prioritization order:
//  1. Dormant projects with uncommitted changes (HasUncommitted=true), oldest LastActive first.
//  2. Any other dormant project (IsDormant=true), oldest LastActive first.
//  3. (Fallback) Any active project with uncommitted changes, oldest LastActive first.
func MostUrgent(reg Registry, threshold time.Duration, now time.Time) (name string, project Project, ok bool) {
	var bestName string
	var bestProj Project
	found := false

	// Phase 1: Dormant projects with uncommitted changes (highest motivational value)
	for k, p := range reg {
		if IsDormant(p, threshold, now) && p.HasUncommitted {
			if !found || p.LastActive.Before(bestProj.LastActive) {
				bestName, bestProj = k, p
				found = true
			}
		}
	}
	if found {
		return bestName, bestProj, true
	}

	// Phase 2: Any other dormant project
	for k, p := range reg {
		if IsDormant(p, threshold, now) {
			if !found || p.LastActive.Before(bestProj.LastActive) {
				bestName, bestProj = k, p
				found = true
			}
		}
	}
	if found {
		return bestName, bestProj, true
	}

	// Phase 3: Any project with uncommitted changes (even if not strictly past threshold)
	for k, p := range reg {
		if p.HasUncommitted {
			if !found || p.LastActive.Before(bestProj.LastActive) {
				bestName, bestProj = k, p
				found = true
			}
		}
	}

	return bestName, bestProj, found
}
