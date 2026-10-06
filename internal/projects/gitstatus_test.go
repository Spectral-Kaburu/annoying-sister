package projects_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Spectral-Kaburu/annoying-sister/internal/projects"
	"github.com/stretchr/testify/require"
)

func TestCheckGitStatus_NonGitDir(t *testing.T) {
	dir := t.TempDir()
	st := projects.CheckGitStatus(dir)
	require.False(t, st.OK)
}

func TestCheckGitStatus_CleanGitRepo(t *testing.T) {
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	file := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(file, []byte("# Test\n"), 0o644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@test.com", "commit", "-m", "initial")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	st := projects.CheckGitStatus(dir)
	require.True(t, st.OK)
	require.False(t, st.HasUncommitted)
	require.Equal(t, 0, st.UncommittedCount)
}

func TestCheckGitStatus_DirtyGitRepo(t *testing.T) {
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	file := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(file, []byte("# Test\n"), 0o644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@test.com", "commit", "-m", "initial")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	// Create 2 untracked / modified files
	require.NoError(t, os.WriteFile(file, []byte("# Modified Test\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0o644))

	st := projects.CheckGitStatus(dir)
	require.True(t, st.OK)
	require.True(t, st.HasUncommitted)
	require.Equal(t, 2, st.UncommittedCount)
}
