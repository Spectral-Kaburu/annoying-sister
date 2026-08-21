package projects

import "time"

// IsDormant reports whether p has gone longer than threshold since
// LastActive, relative to now.
func IsDormant(p Project, threshold time.Duration, now time.Time) bool {
	return now.Sub(p.LastActive) > threshold
}

// MostDormant returns the name and entry of the project with the oldest
// LastActive timestamp in reg, and true if reg is non-empty. Used by the
// on-start trigger to decide which project (if any) to mention.
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
