package provenance

import (
	"fmt"
	"strings"
	"sync"
)

// Store is a durable append-only JSONL ledger. The per-path process lock makes
// separate Store values safe to use concurrently in one process. The manifest
// is a two-phase commit record containing the complete base or committed
// canonical ledger bytes; the JSONL file is a materialized append-only view.
type Store struct {
	path string
	mu   sync.Mutex
}

var pathLocks sync.Map // map[string]*sync.Mutex

// New returns a store backed by path. It creates no files until append.
func New(path string) *Store { return &Store{path: strings.TrimSpace(path)} }

// Open validates an existing ledger before returning a store.
func Open(path string) (*Store, error) {
	store := New(path)
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("provenance store path is required")
	}
	if _, err := store.Read(ReadOptions{}); err != nil {
		return nil, err
	}
	return store, nil
}

// Path returns the configured JSONL path.
func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// Close performs a final validation. Writes are synchronized and closed per
// append, so no process-held file descriptor needs closing.
func (s *Store) Close() error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return fmt.Errorf("provenance store path is required")
	}
	_, err := s.Read(ReadOptions{})
	return err
}

// Append validates and durably appends new events in supplied chain order. A
// missing predecessor is filled from the current tail; a supplied one must
// match exactly.
func (s *Store) Append(records ...Evidence) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return fmt.Errorf("provenance store path is required")
	}
	if len(records) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	lock := lockForPath(s.path)
	lock.Lock()
	defer lock.Unlock()
	return s.appendUnlocked(records)
}
