package projects

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func touchFile(t *testing.T, path string, modTime time.Time) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
	require.NoError(t, os.Chtimes(path, modTime, modTime))
}

// resetAllDirMtimes walks root and sets every directory's own mtime to
// baseline. Creating files/subdirs naturally bumps ancestor directory
// mtimes to "now" as a side effect; tests need to reset that noise to a
// known baseline before asserting on a specific, deliberately newer
// timestamp elsewhere in the tree.
func resetAllDirMtimes(t *testing.T, root string, baseline time.Time) {
	t.Helper()
	require.NoError(t, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if info.IsDir() {
			require.NoError(t, os.Chtimes(path, baseline, baseline))
		}
		return nil
	}))
}

func TestComputeLastActive_FindsDeeplyNestedFile(t *testing.T) {
	root := t.TempDir()
	old := time.Now().Add(-72 * time.Hour)
	recent := time.Now().Add(-1 * time.Hour)

	touchFile(t, filepath.Join(root, "README.md"), old)
	touchFile(t, filepath.Join(root, "src", "nested", "deep", "file.go"), old)

	// Creating the tree bumped every ancestor directory's mtime to "now";
	// pin those back to "old" so only file.go's deliberate timestamp can
	// win.
	resetAllDirMtimes(t, root, old)
	require.NoError(t, os.Chtimes(filepath.Join(root, "src", "nested", "deep", "file.go"), recent, recent))

	got, err := ComputeLastActive(root)
	require.NoError(t, err)
	require.WithinDuration(t, recent, got, time.Second)
}

func TestComputeLastActive_ExcludesDotGit(t *testing.T) {
	root := t.TempDir()
	old := time.Now().Add(-72 * time.Hour)
	veryRecentInGit := time.Now()

	touchFile(t, filepath.Join(root, "README.md"), old)
	// A file inside .git with a much newer mtime (e.g. from git's own
	// bookkeeping) must not count as project activity.
	touchFile(t, filepath.Join(root, ".git", "index"), old)

	resetAllDirMtimes(t, root, old)
	// Only re-bump the .git-internal file/dir mtimes to "now" — these must
	// be excluded from the result.
	require.NoError(t, os.Chtimes(filepath.Join(root, ".git", "index"), veryRecentInGit, veryRecentInGit))
	require.NoError(t, os.Chtimes(filepath.Join(root, ".git"), veryRecentInGit, veryRecentInGit))

	got, err := ComputeLastActive(root)
	require.NoError(t, err)
	require.WithinDuration(t, old, got, time.Second)
}

func TestComputeLastActive_DirectoryMtimeCounts(t *testing.T) {
	root := t.TempDir()
	old := time.Now().Add(-72 * time.Hour)
	touchFile(t, filepath.Join(root, "README.md"), old)
	resetAllDirMtimes(t, root, old)

	// Adding a new (empty) subdirectory changes that directory's own
	// mtime to "now", with no file edit involved at all — this is the
	// case that pure file-mtime tracking would miss.
	before := time.Now()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "newdir"), 0o755))

	got, err := ComputeLastActive(root)
	require.NoError(t, err)
	require.False(t, got.Before(before), "a newly created subdirectory's own mtime must be picked up as activity")
}
