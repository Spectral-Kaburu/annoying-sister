package prompts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Spectral-Kaburu/annoying-sister/internal/prompts"
	"github.com/stretchr/testify/require"
)

func TestEnsureDefaults_SeedsFiles(t *testing.T) {
	tempDir := t.TempDir()

	err := prompts.EnsureDefaults(tempDir)
	require.NoError(t, err)

	for _, name := range prompts.AllFiles {
		p := filepath.Join(tempDir, name)
		info, err := os.Stat(p)
		require.NoError(t, err, "expected file %s to exist", name)
		require.False(t, info.IsDir())
	}
}

func TestEnsureDefaults_DoesNotOverwriteExisting(t *testing.T) {
	tempDir := t.TempDir()
	customContent := "Custom prompt line\n# Comment\n"
	customFile := filepath.Join(tempDir, prompts.FileIdle)
	err := os.WriteFile(customFile, []byte(customContent), 0o644)
	require.NoError(t, err)

	err = prompts.EnsureDefaults(tempDir)
	require.NoError(t, err)

	data, err := os.ReadFile(customFile)
	require.NoError(t, err)
	require.Equal(t, customContent, string(data))
}

func TestStore_FallsBackToEmbedded(t *testing.T) {
	// Empty dir with no files
	tempDir := t.TempDir()
	store := prompts.NewStore(tempDir)

	idle := store.Idle()
	require.NotEmpty(t, idle)
	require.Contains(t, idle[0], "{IdleMinutes}")

	roasts := store.RoastsIdle()
	require.NotEmpty(t, roasts)
	require.Contains(t, roasts[0], "{Player}")
}

func TestStore_ReadsFromDisk(t *testing.T) {
	tempDir := t.TempDir()
	customFile := filepath.Join(tempDir, prompts.FileIdle)
	err := os.WriteFile(customFile, []byte("# Comment\nCustom idle prompt for {IdleMinutes} min.\n\n"), 0o644)
	require.NoError(t, err)

	store := prompts.NewStore(tempDir)
	idle := store.Idle()
	require.Len(t, idle, 1)
	require.Equal(t, "Custom idle prompt for {IdleMinutes} min.", idle[0])
}
