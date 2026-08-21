package projects

import (
	"os"
	"path/filepath"
)

// Discover walks each directory in scanRoots (immediate children only —
// not a recursive search for nested repos within repos) and returns the
// absolute paths of every child directory that contains a .git
// subdirectory, excluding any path present in ignoredPaths.
//
// A scan root that doesn't exist or can't be read is skipped rather than
// treated as a fatal error, so one bad root in config.json doesn't take
// down discovery for the others.
func Discover(scanRoots []string, ignoredPaths []string) ([]string, error) {
	ignored := make(map[string]bool, len(ignoredPaths))
	for _, p := range ignoredPaths {
		ignored[filepath.Clean(p)] = true
	}

	var found []string
	for _, root := range scanRoots {
		entries, err := os.ReadDir(root)
		if err != nil {
			// Missing/unreadable scan root: skip it, not fatal.
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			childPath := filepath.Clean(filepath.Join(root, entry.Name()))
			if ignored[childPath] {
				continue
			}
			gitPath := filepath.Join(childPath, ".git")
			info, err := os.Stat(gitPath)
			if err != nil || !info.IsDir() {
				continue
			}
			found = append(found, childPath)
		}
	}
	return found, nil
}
