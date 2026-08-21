package projects

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixedLastActive(t time.Time) func(string) LastActiveResult {
	return func(string) LastActiveResult {
		return LastActiveResult{LastActive: t, OK: true}
	}
}

func TestMerge_NewDiscoveredDirectory_Inserted(t *testing.T) {
	existing := Registry{}
	discovered := []string{"/home/max/projects/vidking"}
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)

	result := Merge(existing, discovered, fixedLastActive(now))

	require.Len(t, result, 1)
	p, ok := result["vidking"]
	require.True(t, ok, "expected key derived from directory basename")
	assert.Equal(t, "/home/max/projects/vidking", p.Path)
	assert.Equal(t, "", p.Summary)
	assert.Equal(t, SourceAuto, p.Source)
	assert.Equal(t, now, p.LastActive)
}

func TestMerge_MatchingPath_UpdatesPathAndLastActiveOnly_NeverSummary(t *testing.T) {
	existing := Registry{
		"VidKing": Project{
			Path:       "/home/max/projects/vidking",
			Summary:    "my video project, hand-authored",
			LastActive: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:     SourceManual,
		},
	}
	discovered := []string{"/home/max/projects/vidking"}
	newTime := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)

	result := Merge(existing, discovered, fixedLastActive(newTime))

	require.Len(t, result, 1)
	p := result["VidKing"]
	assert.Equal(t, "/home/max/projects/vidking", p.Path)
	assert.Equal(t, newTime, p.LastActive)
	// Summary and Source must be preserved exactly, regardless of Source value.
	assert.Equal(t, "my video project, hand-authored", p.Summary)
	assert.Equal(t, SourceManual, p.Source)
}

func TestMerge_MissingFromScan_NeverDeleted(t *testing.T) {
	existing := Registry{
		"OldProject": Project{
			Path:       "/home/max/archive/old-project",
			Summary:    "not touched in ages",
			LastActive: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:     SourceManual,
		},
	}
	// This scan cycle discovers nothing at all — e.g. scan_roots don't
	// include OldProject's location.
	discovered := []string{}

	result := Merge(existing, discovered, fixedLastActive(time.Now()))

	require.Len(t, result, 1)
	p, ok := result["OldProject"]
	require.True(t, ok, "existing entry must never be deleted just because a scan didn't find it")
	assert.Equal(t, existing["OldProject"], p, "entry must be left completely untouched")
}

func TestMerge_AutoSourceEntry_SummaryStillNeverOverwritten(t *testing.T) {
	// Guards against a subtle bug: "never overwrite summary" must hold
	// regardless of Source, not just for manual entries.
	existing := Registry{
		"vidking": Project{
			Path:       "/home/max/projects/vidking",
			Summary:    "user added a summary to an auto-discovered project",
			LastActive: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:     SourceAuto,
		},
	}
	discovered := []string{"/home/max/projects/vidking"}

	result := Merge(existing, discovered, fixedLastActive(time.Now()))

	assert.Equal(t, "user added a summary to an auto-discovered project", result["vidking"].Summary)
}

func TestMerge_MixedCycle_AllFourBehaviorsAtOnce(t *testing.T) {
	existing := Registry{
		"Known": Project{
			Path:       "/root/known",
			Summary:    "keep me",
			LastActive: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:     SourceAuto,
		},
		"Untouched": Project{
			Path:       "/root/untouched",
			Summary:    "never scanned this cycle",
			LastActive: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:     SourceManual,
		},
	}
	discovered := []string{"/root/known", "/root/brandnew"}
	newTime := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)

	result := Merge(existing, discovered, fixedLastActive(newTime))

	require.Len(t, result, 3)
	assert.Equal(t, newTime, result["Known"].LastActive)
	assert.Equal(t, "keep me", result["Known"].Summary)
	assert.Equal(t, existing["Untouched"], result["Untouched"])
	assert.Equal(t, "/root/brandnew", result["brandnew"].Path)
	assert.Equal(t, SourceAuto, result["brandnew"].Source)
}

func TestMerge_KeyCollision_GetsUniqueSuffix(t *testing.T) {
	existing := Registry{
		"vidking": Project{Path: "/root/a/vidking", Source: SourceAuto},
	}
	// A second scan_root happens to contain a same-named subdirectory.
	discovered := []string{"/root/a/vidking", "/root/b/vidking"}

	result := Merge(existing, discovered, fixedLastActive(time.Now()))

	require.Len(t, result, 2)
	assert.Equal(t, "/root/a/vidking", result["vidking"].Path)
	assert.Equal(t, "/root/b/vidking", result["vidking-2"].Path)
}

func TestMerge_LastActiveComputationFailure_KeepsOldValue(t *testing.T) {
	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := Registry{
		"proj": Project{Path: "/root/proj", LastActive: old, Source: SourceAuto},
	}
	discovered := []string{"/root/proj"}

	failing := func(string) LastActiveResult { return LastActiveResult{OK: false} }
	result := Merge(existing, discovered, failing)

	assert.Equal(t, old, result["proj"].LastActive, "a failed last_active computation should not clobber the existing value")
}
