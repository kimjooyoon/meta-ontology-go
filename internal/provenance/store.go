package provenance

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Store is a durable append-only JSONL ledger. The per-path process lock makes
// separate Store values safe to use concurrently in one process. The manifest
// records the exact byte length and digest last committed.
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

func (s *Store) appendUnlocked(records []Evidence) error {
	state, err := readLedger(s.path)
	if err != nil {
		return err
	}
	known := make(map[string]Evidence, len(state.records)+len(records))
	for _, record := range state.records {
		known[record.ID] = record
	}
	pending := make([]Evidence, 0, len(records))
	for index, input := range records {
		normalized, err := normalizeEvidence(input)
		if err != nil {
			return fmt.Errorf("event %d: %w", index, err)
		}
		if existing, exists := known[normalized.ID]; exists {
			candidate, err := materializeDuplicate(normalized, existing)
			if err != nil {
				return fmt.Errorf("event %d: %w", index, err)
			}
			left, err := marshalEvidence(existing, true)
			if err != nil {
				return err
			}
			right, err := marshalEvidence(candidate, true)
			if err != nil {
				return err
			}
			if !bytes.Equal(left, right) {
				return &ConflictError{ID: normalized.ID, Detail: "same event ID has different canonical bytes"}
			}
			continue
		}
		sequence := uint64(len(state.records) + len(pending) + 1)
		previous := lastPending(state.records, pending)
		if normalized.Sequence != 0 && normalized.Sequence != sequence {
			return &ChainError{ID: normalized.ID, Expected: fmt.Sprint(sequence), Actual: fmt.Sprint(normalized.Sequence), Detail: "sequence is not contiguous"}
		}
		normalized.Sequence = sequence
		if previous == nil {
			if normalized.Predecessor != nil {
				return &ChainError{ID: normalized.ID, Detail: "first event cannot claim a predecessor"}
			}
		} else {
			expected := &DigestLink{ID: previous.ID, Digest: previous.Hash}
			if normalized.Predecessor == nil {
				normalized.Predecessor = expected
			} else if normalized.Predecessor.ID != expected.ID || normalized.Predecessor.Digest != expected.Digest {
				return &ChainError{ID: normalized.ID, Expected: expected.ID + "/" + expected.Digest, Actual: normalized.Predecessor.ID + "/" + normalized.Predecessor.Digest, Detail: "predecessor does not identify the current tail"}
			}
		}
		normalized, err = finishRecord(normalized)
		if err != nil {
			return fmt.Errorf("event %d: %w", index, err)
		}
		known[normalized.ID] = normalized
		pending = append(pending, normalized)
	}
	if len(pending) == 0 {
		return nil
	}
	payload, err := canonicalJSONL(pending)
	if err != nil {
		return err
	}
	if err := appendPayload(s.path, payload); err != nil {
		return err
	}
	combined := append(append([]byte(nil), state.bytes...), payload...)
	allRecords := append(append([]Evidence(nil), state.records...), pending...)
	return writeManifest(s.path, ledgerManifestFor(combined, allRecords))
}

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

func canonicalJSONL(records []Evidence) ([]byte, error) {
	var output bytes.Buffer
	for _, record := range records {
		line, err := marshalEvidence(record, true)
		if err != nil {
			return nil, err
		}
		output.Write(line)
		output.WriteByte('\n')
	}
	return output.Bytes(), nil
}

func appendPayload(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create provenance directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open provenance store: %w", err)
	}
	if err := writeFull(file, payload); err != nil {
		_ = file.Close()
		return fmt.Errorf("append provenance evidence: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync provenance store: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close provenance store: %w", err)
	}
	return nil
}

func writeFull(file *os.File, payload []byte) error {
	for len(payload) > 0 {
		written, err := file.Write(payload)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		payload = payload[written:]
	}
	return nil
}

// ReadAll is a convenience alias for adapters that prefer explicit wording.
func (s *Store) ReadAll(options ReadOptions) (Snapshot, error) { return s.Read(options) }
