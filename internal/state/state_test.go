package state

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_MissingFile_StartsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := Load(path)
	require.NoError(t, err)
	assert.True(t, s.IdleLastNudged().IsZero())
	_, ok := s.ProjectLastNudgedDate("anything")
	assert.False(t, ok)
}

func TestSetIdleLastNudged_PersistsAcrossReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := Load(path)
	require.NoError(t, err)

	ts := time.Date(2026, 8, 20, 9, 14, 0, 0, time.UTC)
	require.NoError(t, s.SetIdleLastNudged(ts))

	reloaded, err := Load(path)
	require.NoError(t, err)
	assert.True(t, ts.Equal(reloaded.IdleLastNudged()))
}

func TestSetProjectLastNudgedDate_PersistsAndIsDateOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := Load(path)
	require.NoError(t, err)

	require.NoError(t, s.SetProjectLastNudgedDate("VidKing", "2026-08-19"))

	reloaded, err := Load(path)
	require.NoError(t, err)
	date, ok := reloaded.ProjectLastNudgedDate("VidKing")
	require.True(t, ok)
	assert.Equal(t, "2026-08-19", date)
}

func TestSetProjectLastNudgedDate_DoesNotAffectOtherProjects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := Load(path)
	require.NoError(t, err)

	require.NoError(t, s.SetProjectLastNudgedDate("A", "2026-08-18"))
	require.NoError(t, s.SetProjectLastNudgedDate("B", "2026-08-19"))

	dateA, _ := s.ProjectLastNudgedDate("A")
	dateB, _ := s.ProjectLastNudgedDate("B")
	assert.Equal(t, "2026-08-18", dateA)
	assert.Equal(t, "2026-08-19", dateB)
}
