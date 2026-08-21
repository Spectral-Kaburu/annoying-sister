package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_MissingFile_CreatesDefaultsAndPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, Default(), cfg)

	// Must actually have been written to disk, not just returned in memory.
	_, statErr := os.Stat(path)
	require.NoError(t, statErr)

	// A second load should read back the same values from disk.
	cfg2, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, cfg, cfg2)
}

func TestLoad_ExistingFile_RespectsUserEdits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	custom := `{
		"scan_roots": ["/home/max/code"],
		"ignored_paths": ["/home/max/code/vendor"],
		"idle_threshold_minutes": 5,
		"idle_repeat_interval_minutes": 10,
		"idle_poll_interval_seconds": 15,
		"dormant_threshold_days": 3,
		"project_scan_interval_minutes": 30,
		"spectretts_socket_path": "/tmp/custom.sock"
	}`
	require.NoError(t, os.WriteFile(path, []byte(custom), 0o644))

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"/home/max/code"}, cfg.ScanRoots)
	assert.Equal(t, 5, cfg.IdleThresholdMinutes)
	assert.Equal(t, "/tmp/custom.sock", cfg.SpectreTTSSocketPath)
}
