package projects_test

import (
	"testing"
	"time"

	"github.com/Spectral-Kaburu/annoying-sister/internal/projects"
	"github.com/stretchr/testify/require"
)

func TestMostUrgent_PrioritizesDormantWithUncommitted(t *testing.T) {
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	threshold := 7 * 24 * time.Hour

	reg := projects.Registry{
		"CleanOld": projects.Project{
			Path:           "/clean",
			LastActive:     now.Add(-20 * 24 * time.Hour), // 20 days ago, clean
			HasUncommitted: false,
		},
		"DirtyDormant": projects.Project{
			Path:             "/dirty",
			LastActive:       now.Add(-10 * 24 * time.Hour), // 10 days ago, dirty
			HasUncommitted:   true,
			UncommittedCount: 3,
		},
		"DirtyRecent": projects.Project{
			Path:             "/recent",
			LastActive:       now.Add(-2 * 24 * time.Hour), // 2 days ago, dirty
			HasUncommitted:   true,
			UncommittedCount: 1,
		},
	}

	name, proj, ok := projects.MostUrgent(reg, threshold, now)
	require.True(t, ok)
	require.Equal(t, "DirtyDormant", name)
	require.True(t, proj.HasUncommitted)
}

func TestMostUrgent_FallsBackToCleanDormant(t *testing.T) {
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	threshold := 7 * 24 * time.Hour

	reg := projects.Registry{
		"CleanOld": projects.Project{
			Path:           "/clean_old",
			LastActive:     now.Add(-30 * 24 * time.Hour),
			HasUncommitted: false,
		},
		"CleanNewer": projects.Project{
			Path:           "/clean_newer",
			LastActive:     now.Add(-10 * 24 * time.Hour),
			HasUncommitted: false,
		},
	}

	name, _, ok := projects.MostUrgent(reg, threshold, now)
	require.True(t, ok)
	require.Equal(t, "CleanOld", name)
}
