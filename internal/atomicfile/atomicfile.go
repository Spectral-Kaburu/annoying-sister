// Package atomicfile provides a small helper for writing files atomically.
//
// This is not for concurrent-writer protection — Nudge is the sole owner
// and sole writer of every file it touches, so there is never a second
// writer to race against. It exists purely as basic hygiene against a
// crash or power loss leaving a truncated/corrupt file mid-write: the new
// content is written to a temp file in the same directory, then swapped
// into place with a single os.Rename, which is atomic on the same
// filesystem.
package atomicfile

import (
	"os"
	"path/filepath"
)

// Write atomically replaces the file at path with data. The temp file is
// created in the same directory as path so the final os.Rename is a same-
// filesystem rename (and therefore atomic).
func Write(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	// Clean up the temp file if we return before the rename succeeds.
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	success = true
	return nil
}
