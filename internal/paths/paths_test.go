package paths_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Spectral-Kaburu/annoying-sister/internal/paths"
	"github.com/stretchr/testify/require"
)

func TestPaths(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	expectedDir := filepath.Join(tempHome, ".blackboxx", "nudge")

	dir, err := paths.NudgeDir()
	require.NoError(t, err)
	require.Equal(t, expectedDir, dir)

	cfg, err := paths.ConfigPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(expectedDir, "config.json"), cfg)

	proj, err := paths.ProjectsPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(expectedDir, "projects.json"), proj)

	state, err := paths.StatePath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(expectedDir, "state.json"), state)
}

func TestEnsureNudgeDir(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	expectedDir := filepath.Join(tempHome, ".blackboxx", "nudge")

	err := paths.EnsureNudgeDir()
	require.NoError(t, err)

	info, err := os.Stat(expectedDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}
