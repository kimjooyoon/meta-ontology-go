package provenance

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Read validates the complete physical ledger and returns append order.
func (s *Store) Read(options ReadOptions) (Snapshot, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return Snapshot{}, fmt.Errorf("provenance store path is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	lock := lockForPath(s.path)
	lock.Lock()
	defer lock.Unlock()
	state, err := readLedger(s.path)
	if err != nil {
		return Snapshot{}, err
	}
	if err := checkReadOptions(state.records, options); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Records: state.records, Digest: state.digest}, nil
}

// Verify applies explicit verified claims. Candidate and deferred records
// remain readable evidence but cannot satisfy these claims.
func (s *Store) Verify(claims ...VerifiedClaim) (Snapshot, error) {
	return s.Read(ReadOptions{RequiredVerified: claims})
}
func (s *Store) appendRecords(records []Evidence) error { return s.Append(records...) }
func lockForPath(path string) *sync.Mutex {
	absolute, err := filepath.Abs(path)
	if err != nil {
		absolute = path
	}
	value, _ := pathLocks.LoadOrStore(absolute, &sync.Mutex{})
	return value.(*sync.Mutex)
}
func lastPending(existing, pending []Evidence) *Evidence {
	if len(pending) > 0 {
		return &pending[len(pending)-1]
	}
	if len(existing) > 0 {
		return &existing[len(existing)-1]
	}
	return nil
}
func materializeDuplicate(input, existing Evidence) (Evidence, error) {
	if input.Sequence == 0 {
		input.Sequence = existing.Sequence
	}
	if input.Predecessor == nil {
		input.Predecessor = existing.Predecessor
	}
	return finishRecord(input)
}
func finishRecord(record Evidence) (Evidence, error) {
	unsigned, err := marshalEvidence(record, false)
	if err != nil {
		return Evidence{}, err
	}
	expected := digestBytes(unsigned)
	if record.Hash != "" && strings.ToLower(strings.TrimSpace(record.Hash)) != expected {
		return Evidence{}, &ConflictError{ID: record.ID, Field: "hash", Detail: fmt.Sprintf("expected %q, got %q", expected, record.Hash)}
	}
	record.Hash = expected
	return record, nil
}
