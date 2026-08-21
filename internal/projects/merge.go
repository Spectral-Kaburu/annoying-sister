package projects

import (
	"path/filepath"
	"strconv"
	"time"
)

// LastActiveResult lets the injected last-active computation report
// failure (e.g. the directory vanished mid-scan) explicitly, rather than
// Merge having to guess whether a zero time.Time means "really zero" or
// "computation failed".
type LastActiveResult struct {
	LastActive time.Time
	OK         bool
}

// Merge folds a fresh discovery pass into the existing registry. It
// implements exactly the four cases from the spec's "Merge behavior"
// section:
//
//  1. New discovered directory, no existing entry (matched by Path, not by
//     map key): insert with Source: SourceAuto, empty Summary, computed
//     LastActive.
//  2. Discovered directory matches an existing entry by Path: update Path
//     and LastActive only. Summary is never touched, regardless of
//     Source.
//  3. Existing entry whose directory was not found this cycle: left
//     completely untouched. Absence from a scan is never a deletion
//     signal.
//  4. (Implicit) An existing entry's Source and Summary are preserved
//     across every update — only Path and LastActive ever change for an
//     already-known project.
//
// Matching is by Path, not by registry key, because a hand-added entry
// may use any display name while its Path is what discovery actually
// finds again next cycle.
//
// computeLastActive is injected so Merge itself never touches the
// filesystem, keeping this highest-risk piece of logic unit-testable with
// plain in-memory arguments.
func Merge(existing Registry, discoveredPaths []string, computeLastActive func(path string) LastActiveResult) Registry {
	// Index existing entries by their current path so we can find a match
	// regardless of what key/name they're stored under.
	pathToKey := make(map[string]string, len(existing))
	for key, p := range existing {
		pathToKey[filepath.Clean(p.Path)] = key
	}

	result := make(Registry, len(existing)+len(discoveredPaths))
	for k, v := range existing {
		result[k] = v // start from a full copy; case 3 needs no further action
	}

	usedKeys := make(map[string]bool, len(result))
	for k := range result {
		usedKeys[k] = true
	}

	for _, rawPath := range discoveredPaths {
		path := filepath.Clean(rawPath)

		if key, ok := pathToKey[path]; ok {
			// Case 2: known project, update Path + LastActive only.
			entry := result[key]
			entry.Path = path
			if r := computeLastActive(path); r.OK {
				entry.LastActive = r.LastActive
			}
			result[key] = entry
			continue
		}

		// Case 1: brand new project.
		key := uniqueKey(usedKeys, filepath.Base(path))
		usedKeys[key] = true

		entry := Project{
			Path:    path,
			Summary: "",
			Source:  SourceAuto,
		}
		if r := computeLastActive(path); r.OK {
			entry.LastActive = r.LastActive
		}
		result[key] = entry
	}

	// Case 3 requires no code: entries whose path wasn't in
	// discoveredPaths were already copied into result above and are never
	// touched again in this function.

	return result
}

// uniqueKey returns base if it isn't already taken in used, otherwise
// appends "-2", "-3", ... until it finds a free one. This only matters if
// two different scan_roots happen to contain same-named subdirectories.
func uniqueKey(used map[string]bool, base string) string {
	if !used[base] {
		return base
	}
	for i := 2; ; i++ {
		candidate := base + "-" + strconv.Itoa(i)
		if !used[candidate] {
			return candidate
		}
	}
}
