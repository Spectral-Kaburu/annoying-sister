// Package projects implements the project registry (projects.json):
// discovery of git-containing directories under scan_roots, the
// last_active activity signal, and the merge logic that lets
// scanner-discovered and hand-edited entries coexist safely.
package projects

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/Spectral-Kaburu/annoying-sister/internal/atomicfile"
)

// Project is one entry in projects.json.
type Project struct {
	Path             string    `json:"path"`
	Summary          string    `json:"summary"`
	LastActive       time.Time `json:"last_active"`
	Source           string    `json:"source"` // "auto" | "manual"
	HasUncommitted   bool      `json:"has_uncommitted,omitempty"`
	UncommittedCount int       `json:"uncommitted_count,omitempty"`
}

const (
	SourceAuto   = "auto"
	SourceManual = "manual"
)

// Registry is the full projects.json document: project name -> Project.
// The map key doubles as the spoken ProjectName in nudge templates.
type Registry map[string]Project

// Store is the in-memory, mutex-guarded view of projects.json, shared
// between the project scanner goroutine (sole writer) and the on-start /
// dormant-nudge decision logic (readers).
type Store struct {
	mu   sync.Mutex
	path string
	reg  Registry
}

// Load reads projects.json from path. A missing file is treated as an
// empty registry rather than an error — expected on first run.
func Load(path string) (*Store, error) {
	s := &Store{path: path, reg: make(Registry)}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	if reg == nil {
		reg = make(Registry)
	}
	s.reg = reg
	return s, nil
}

// Snapshot returns a copy of the current registry, safe for the caller to
// read without further locking.
func (s *Store) Snapshot() Registry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(Registry, len(s.reg))
	for k, v := range s.reg {
		out[k] = v
	}
	return out
}

// Replace atomically swaps in a newly merged registry and persists it to
// disk. This is the only way the registry is mutated — always via a full
// merge result from the scanner, never a partial in-place edit.
func (s *Store) Replace(newReg Registry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reg = newReg
	return s.flushLocked()
}

func (s *Store) flushLocked() error {
	data, err := json.MarshalIndent(s.reg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return atomicfile.Write(s.path, data, 0o644)
}
