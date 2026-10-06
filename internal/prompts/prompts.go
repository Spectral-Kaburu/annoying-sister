package prompts

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed defaults/*.txt
var defaultFS embed.FS

const (
	FileIdle               = "idle.txt"
	FileDormant            = "dormant.txt"
	FileDormantUncommitted = "dormant_uncommitted.txt"
	FileStartupGeneric     = "startup_generic.txt"
	FileStartupDormant     = "startup_dormant.txt"
	FileStartupUncommitted = "startup_uncommitted.txt"
	FileRoastsIdle         = "roasts_idle.txt"
	FileRoastsUncommitted  = "roasts_uncommitted.txt"
	FileRoastsDormant      = "roasts_dormant.txt"
)

// AllFiles lists all prompt filenames managed by Nudge.
var AllFiles = []string{
	FileIdle,
	FileDormant,
	FileDormantUncommitted,
	FileStartupGeneric,
	FileStartupDormant,
	FileStartupUncommitted,
	FileRoastsIdle,
	FileRoastsUncommitted,
	FileRoastsDormant,
}

// Store loads and parses prompts from a directory, falling back to embedded defaults.
type Store struct {
	dir string
	mu  sync.RWMutex
}

// NewStore creates a prompt Store pointing to the given directory.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// EnsureDefaults seeds missing prompt files into dir from embedded defaults.
func EnsureDefaults(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create prompts dir: %w", err)
	}

	entries, err := fs.ReadDir(defaultFS, "defaults")
	if err != nil {
		return fmt.Errorf("read embedded defaults: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		dest := filepath.Join(dir, name)

		if _, err := os.Stat(dest); err == nil {
			// File already exists; do not overwrite user modifications.
			continue
		}

		data, err := defaultFS.ReadFile("defaults/" + name)
		if err != nil {
			return fmt.Errorf("read embedded file %s: %w", name, err)
		}

		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write prompt file %s: %w", dest, err)
		}
	}
	return nil
}

// parsePromptLines parses raw text into a slice of non-empty prompt strings,
// skipping comments (lines starting with #) and blank lines.
func parsePromptLines(content string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

// GetPool returns the prompt lines for a given filename. It first checks
// the disk directory; if missing, empty, or unreadable, it falls back to embedded defaults.
func (s *Store) GetPool(filename string) []string {
	if s != nil && s.dir != "" {
		diskPath := filepath.Join(s.dir, filename)
		if data, err := os.ReadFile(diskPath); err == nil {
			lines := parsePromptLines(string(data))
			if len(lines) > 0 {
				return lines
			}
		}
	}

	// Fallback to embedded default
	data, err := defaultFS.ReadFile("defaults/" + filename)
	if err == nil {
		lines := parsePromptLines(string(data))
		if len(lines) > 0 {
			return lines
		}
	}

	return nil
}

func (s *Store) Idle() []string {
	return s.GetPool(FileIdle)
}

func (s *Store) Dormant() []string {
	return s.GetPool(FileDormant)
}

func (s *Store) DormantUncommitted() []string {
	return s.GetPool(FileDormantUncommitted)
}

func (s *Store) StartupGeneric() []string {
	return s.GetPool(FileStartupGeneric)
}

func (s *Store) StartupDormant() []string {
	return s.GetPool(FileStartupDormant)
}

func (s *Store) StartupUncommitted() []string {
	return s.GetPool(FileStartupUncommitted)
}

func (s *Store) RoastsIdle() []string {
	return s.GetPool(FileRoastsIdle)
}

func (s *Store) RoastsUncommitted() []string {
	return s.GetPool(FileRoastsUncommitted)
}

func (s *Store) RoastsDormant() []string {
	return s.GetPool(FileRoastsDormant)
}
