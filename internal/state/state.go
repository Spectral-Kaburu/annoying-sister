// Package state manages Nudge's cooldown bookkeeping (state.json):
// idle.last_nudged and, per project, last_nudged_date.
//
// state.json is written by both the idle watcher and the project scanner
// goroutines, but both are part of the same process and only one of them
// writes to a given top-level key ("idle" vs "projects") in practice. Per
// the spec, a simple in-memory sync.Mutex guarding all reads/writes is
// sufficient — no OS-level file locking is needed, since there is exactly
// one process and Nudge is the sole owner of this file.
package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/maximus/nudge/internal/atomicfile"
)

// IdleState holds the idle-category cooldown timestamp.
type IdleState struct {
	LastNudged time.Time `json:"last_nudged"`
}

// ProjectState holds the dormant-project cooldown date for one project.
type ProjectState struct {
	LastNudgedDate string `json:"last_nudged_date"` // "YYYY-MM-DD", date-only by design
}

// document is the on-disk shape of state.json.
type document struct {
	Idle     IdleState               `json:"idle"`
	Projects map[string]ProjectState `json:"projects"`
}

// Store is the in-memory, mutex-guarded view of state.json. Every mutating
// method flushes the full document to disk (atomically) before returning,
// so state.json never lags what's held in memory.
type Store struct {
	mu   sync.Mutex
	path string
	doc  document
}

// Load reads state.json from path. A missing file is treated as an empty,
// fresh state rather than an error — this is expected on first run.
func Load(path string) (*Store, error) {
	s := &Store{
		path: path,
		doc: document{
			Projects: make(map[string]ProjectState),
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Projects == nil {
		doc.Projects = make(map[string]ProjectState)
	}
	s.doc = doc
	return s, nil
}

// IdleLastNudged returns the timestamp of the last idle nudge. The zero
// Time is returned if none has ever fired.
func (s *Store) IdleLastNudged() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.doc.Idle.LastNudged
}

// SetIdleLastNudged records t as the idle category's last-nudged timestamp
// and flushes state.json to disk.
func (s *Store) SetIdleLastNudged(t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doc.Idle.LastNudged = t
	return s.flushLocked()
}

// ProjectLastNudgedDate returns the stored last_nudged_date for the named
// project ("" if never nudged) and whether an entry exists at all.
func (s *Store) ProjectLastNudgedDate(project string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps, ok := s.doc.Projects[project]
	if !ok {
		return "", false
	}
	return ps.LastNudgedDate, true
}

// SetProjectLastNudgedDate records date (format "YYYY-MM-DD") as the
// dormant-project cooldown date for the named project and flushes
// state.json to disk.
func (s *Store) SetProjectLastNudgedDate(project string, date string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doc.Projects[project] = ProjectState{LastNudgedDate: date}
	return s.flushLocked()
}

// flushLocked serializes the current document and writes it atomically.
// Caller must hold s.mu.
func (s *Store) flushLocked() error {
	data, err := json.MarshalIndent(s.doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return atomicfile.Write(s.path, data, 0o644)
}
