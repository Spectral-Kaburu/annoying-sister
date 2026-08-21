package projects

import (
	"io/fs"
	"path/filepath"
	"time"
)

// ComputeLastActive returns the maximum mtime across every file and
// directory in root, recursively, excluding everything under .git/.
//
// This is deliberately not git-commit-date-based and not top-level- or
// directory-mtime-only: see the spec's "last_active computation" section
// for the full rationale. In short, only a full recursive file+directory
// mtime walk reliably captures "something was edited and saved" at any
// depth, including in-place edits that don't touch a parent directory's
// mtime.
//
// Unreadable entries (permission errors, races with concurrent deletion,
// etc.) are skipped rather than aborting the whole walk — this signal is
// allowed to be noisy/imprecise per the spec ("dormancy nudges tolerate
// being off by a day or two").
func ComputeLastActive(root string) (time.Time, error) {
	var maxTime time.Time

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}
		info, err := d.Info()
		if err != nil {
			// Vanished/unreadable between readdir and stat; skip.
			return nil
		}
		if info.ModTime().After(maxTime) {
			maxTime = info.ModTime()
		}
		return nil
	})

	return maxTime, err
}
