// Package config loads Nudge's user-adjustable settings from config.json.
// All thresholds, intervals, and paths live here rather than as hardcoded
// constants, per the spec. A missing file is created with sensible
// defaults on first run rather than causing a crash.
package config

import (
	"encoding/json"
	"os"

	"github.com/Spectral-Kaburu/annoying-sister/internal/atomicfile"
)

// Config mirrors the config.json schema.
type Config struct {
	ScanRoots                  []string `json:"scan_roots"`
	IgnoredPaths               []string `json:"ignored_paths"`
	IdleThresholdMinutes       int      `json:"idle_threshold_minutes"`
	IdleRepeatIntervalMinutes  int      `json:"idle_repeat_interval_minutes"`
	IdlePollIntervalSeconds    int      `json:"idle_poll_interval_seconds"`
	DormantThresholdDays       int      `json:"dormant_threshold_days"`
	ProjectScanIntervalMinutes int      `json:"project_scan_interval_minutes"`
	SpectreTTSSocketPath       string   `json:"spectretts_socket_path"`
}

// Default returns the built-in defaults used when config.json does not yet
// exist. scan_roots is intentionally left empty in the built-in default —
// the example in the spec hardcodes a single-user path
// (/home/maximus/projects) that would be wrong for anyone else; a first
// run with no scan_roots configured simply discovers no projects until the
// user edits config.json, rather than silently scanning a path that isn't
// theirs.
func Default() Config {
	return Config{
		ScanRoots:                  []string{},
		IgnoredPaths:               []string{},
		IdleThresholdMinutes:       20,
		IdleRepeatIntervalMinutes:  20,
		IdlePollIntervalSeconds:    30,
		DormantThresholdDays:       7,
		ProjectScanIntervalMinutes: 15,
		SpectreTTSSocketPath:       "/tmp/spectretts.sock",
	}
}

// Load reads config.json from path. If the file does not exist, it is
// created with Default() values (atomically) and the defaults are
// returned. Any other read/parse error is returned as-is.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if werr := save(path, cfg); werr != nil {
				return Config{}, werr
			}
			return cfg, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func save(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return atomicfile.Write(path, data, 0o644)
}

// Save writes cfg to path atomically. Public counterpart of save(),
// exposed for the interactive setup wizard (nudged -setup).
func Save(path string, cfg Config) error {
	return save(path, cfg)
}
