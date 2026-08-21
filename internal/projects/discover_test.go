package projects

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mkGitRepo(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
}

func TestDiscover_FindsImmediateGitRepos(t *testing.T) {
	root := t.TempDir()
	mkGitRepo(t, filepath.Join(root, "repo-a"))
	mkGitRepo(t, filepath.Join(root, "repo-b"))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "not-a-repo"), 0o755))

	found, err := Discover([]string{root}, nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		filepath.Join(root, "repo-a"),
		filepath.Join(root, "repo-b"),
	}, found)
}

func TestDiscover_DoesNotRecurseIntoNestedRepos(t *testing.T) {
	root := t.TempDir()
	outer := filepath.Join(root, "outer")
	mkGitRepo(t, outer)
	mkGitRepo(t, filepath.Join(outer, "nested"))

	found, err := Discover([]string{root}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{outer}, found)
}

func TestDiscover_RespectsIgnoredPaths(t *testing.T) {
	root := t.TempDir()
	keep := filepath.Join(root, "keep")
	skip := filepath.Join(root, "skip")
	mkGitRepo(t, keep)
	mkGitRepo(t, skip)

	found, err := Discover([]string{root}, []string{skip})
	require.NoError(t, err)
	assert.Equal(t, []string{keep}, found)
}

func TestDiscover_MissingScanRoot_SkippedNotFatal(t *testing.T) {
	found, err := Discover([]string{"/does/not/exist"}, nil)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestDiscover_MultipleScanRoots(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	mkGitRepo(t, filepath.Join(rootA, "a"))
	mkGitRepo(t, filepath.Join(rootB, "b"))

	found, err := Discover([]string{rootA, rootB}, nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		filepath.Join(rootA, "a"),
		filepath.Join(rootB, "b"),
	}, found)
}
