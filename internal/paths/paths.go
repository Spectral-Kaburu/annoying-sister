// Package paths defines Nudge's on-disk layout. DataRoot is the single
// named constant for the project's central data directory root, per the
// spec's requirement that a future rename of "~/.blackboxx/" be a one-line
// change rather than a repeated string literal scattered across the code.
package paths

import (
	"os"
	"path/filepath"
)

// DataRoot is the name of the project's central data directory, rooted at
// the user's home directory (i.e. ~/.blackboxx/). This is expected to
// become a shared convention across future Aether components, but that is
// not finalized yet — Nudge only depends on its own "nudge" subdirectory
// beneath it. Change this constant, and only this constant, if the root
// directory name is ever finalized differently.
const DataRoot = ".blackboxx"

// serviceDir is Nudge's own subdirectory name beneath DataRoot.
const serviceDir = "nudge"

const (
	configFileName   = "config.json"
	projectsFileName = "projects.json"
	stateFileName    = "state.json"
)

// NudgeDir returns the absolute path to ~/.blackboxx/nudge/.
func NudgeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, DataRoot, serviceDir), nil
}

// ConfigPath returns the absolute path to config.json.
func ConfigPath() (string, error) {
	dir, err := NudgeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// ProjectsPath returns the absolute path to projects.json.
func ProjectsPath() (string, error) {
	dir, err := NudgeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, projectsFileName), nil
}

// StatePath returns the absolute path to state.json.
func StatePath() (string, error) {
	dir, err := NudgeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFileName), nil
}

// EnsureNudgeDir creates ~/.blackboxx/nudge/ (and any missing parents) if
// it does not already exist. Safe to call unconditionally at startup.
func EnsureNudgeDir() error {
	dir, err := NudgeDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}
