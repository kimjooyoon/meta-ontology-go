package provenance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Store is a durable append-only JSONL ledger. The per-path process lock makes
// separate Store values safe to use concurrently in one process. The manifest
// is an integrity index: it records the exact byte length and digest last
// committed, allowing a closed/reopened store to reject tail truncation and
// physical reordering rather than silently accepting a valid prefix.
type Store struct {
	path string
	mu   sync.Mutex
}

var pathLocks sync.Map // map[string]*sync.Mutex

type ledgerState struct {
	records []Evidence
	bytes   []byte
	digest  string
	lines   int
}

type ledgerManifest struct {
	Schema   int    `json:"schema"`
	Bytes    int64  `json:"bytes"`
	Lines    int    `json:"lines"`
	Digest   string `json:"digest"`
	LastID   string `json:"last_id,omitempty"`
	LastHash string `json:"last_hash,omitempty"`
}

// New returns a store backed by path. It does not create the ledger until a
// successful append.
func New(path string) *Store { return &Store{path: strings.TrimSpace(path)} }

// Open validates an existing ledger before returning a store. A missing path
// is a valid empty store; the first append creates it.
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

// Close performs a final validation. Writes are opened, synchronized, and
// closed per append, so no process-held file descriptor needs closing.
func (s *Store) Close() error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return fmt.Errorf("provenance store path is required")
	}
	_, err := s.Read(ReadOptions{})
	return err
}

// Append validates and durably appends new events in the supplied chain order.
// A record with an existing ID is idempotent only when its complete canonical
// bytes match. A missing predecessor is filled from the current tail so a
// producer need not predict the prior content digest; a supplied predecessor
// must still match exactly.
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
	combined := make([]byte, 0, len(state.bytes)+len(payload))
	combined = append(combined, state.bytes...)
	combined = append(combined, payload...)
	return writeManifest(s.path, ledgerManifestFor(combined, append(append([]Evidence(nil), state.records...), pending...)))
}

// Read validates the complete physical ledger and returns it in append order.
// It never sorts records: chain order and raw bytes are integrity-relevant.
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

// Verify reads and applies explicit verified claims. Candidate and deferred
// records remain readable evidence but cannot satisfy these claims.
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

func readLedger(path string) (ledgerState, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if _, manifestErr := os.Stat(manifestPath(path)); manifestErr == nil {
			return ledgerState{}, corruption(path, 0, 0, "ledger-mutation", fmt.Errorf("manifest exists but ledger is missing"))
		}
		return ledgerState{digest: digestBytes(nil)}, nil
	}
	if err != nil {
		return ledgerState{}, fmt.Errorf("open provenance store: %w", err)
	}
	state := ledgerState{bytes: data, digest: digestBytes(data)}
	if len(data) == 0 {
		state.lines = 0
		if err := verifyManifest(path, state); err != nil {
			return ledgerState{}, err
		}
		return state, nil
	}
	if data[len(data)-1] != '\n' {
		return ledgerState{}, corruption(path, 0, int64(len(data)), "truncated", fmt.Errorf("ledger must end with LF"))
	}
	for offset, lineNumber, start := int64(0), 1, 0; start < len(data); lineNumber++ {
		end := bytes.IndexByte(data[start:], '\n')
		if end < 0 {
			return ledgerState{}, corruption(path, lineNumber, offset, "truncated", fmt.Errorf("record has no terminating LF"))
		}
		end += start
		raw := data[start:end]
		if len(raw) == 0 {
			return ledgerState{}, corruption(path, lineNumber, offset, "blank-line", fmt.Errorf("blank lines are not valid evidence records"))
		}
		record, parseErr := parseLine(path, lineNumber, offset, raw)
		if parseErr != nil {
			return ledgerState{}, parseErr
		}
		if record.Sequence != uint64(lineNumber) {
			return ledgerState{}, corruption(path, lineNumber, offset, "chain-gap", fmt.Errorf("sequence %d is not %d", record.Sequence, lineNumber))
		}
		if len(state.records) > 0 {
			previous := state.records[len(state.records)-1]
			expected := &DigestLink{ID: previous.ID, Digest: previous.Hash}
			if record.Predecessor == nil || record.Predecessor.ID != expected.ID || record.Predecessor.Digest != expected.Digest {
				return ledgerState{}, corruption(path, lineNumber, offset, "chain-gap", fmt.Errorf("predecessor does not identify line %d", lineNumber-1))
			}
		} else if record.Predecessor != nil {
			return ledgerState{}, corruption(path, lineNumber, offset, "chain-gap", fmt.Errorf("first record has a predecessor"))
		}
		for _, previous := range state.records {
			if previous.ID == record.ID {
				return ledgerState{}, corruption(path, lineNumber, offset, "duplicate-id", fmt.Errorf("event ID %q already appeared", record.ID))
			}
		}
		state.records = append(state.records, record)
		state.lines++
		start = end + 1
		offset = int64(start)
	}
	if err := verifyManifest(path, state); err != nil {
		return ledgerState{}, err
	}
	return state, nil
}

func parseLine(path string, lineNumber int, offset int64, raw []byte) (Evidence, error) {
	record, err := decodeEvidence(raw)
	if err != nil {
		return Evidence{}, lineDiagnostic(path, lineNumber, offset, err)
	}
	normalized, err := normalizeEvidence(record)
	if err != nil {
		return Evidence{}, corruption(path, lineNumber, offset, "invalid-record", err)
	}
	if record.Hash == "" {
		return Evidence{}, corruption(path, lineNumber, offset, "malformed", fmt.Errorf("hash is required"))
	}
	unsigned, err := marshalEvidence(normalized, false)
	if err != nil {
		return Evidence{}, corruption(path, lineNumber, offset, "invalid-record", err)
	}
	expectedHash := digestBytes(unsigned)
	if strings.ToLower(strings.TrimSpace(record.Hash)) != expectedHash {
		return Evidence{}, corruption(path, lineNumber, offset, "hash-mismatch", fmt.Errorf("expected %q, got %q", expectedHash, record.Hash))
	}
	normalized.Hash = expectedHash
	expectedLine, err := marshalEvidence(normalized, true)
	if err != nil {
		return Evidence{}, corruption(path, lineNumber, offset, "invalid-record", err)
	}
	if !bytes.Equal(raw, expectedLine) {
		return Evidence{}, corruption(path, lineNumber, offset, "non-canonical", fmt.Errorf("line does not match canonical JSON encoding"))
	}
	return normalized, nil
}

func lineDiagnostic(path string, lineNumber int, offset int64, err error) error {
	var detail *lineError
	if errors.As(err, &detail) {
		return corruption(path, lineNumber, offset, detail.kind, detail.err)
	}
	return corruption(path, lineNumber, offset, "malformed", err)
}

func corruption(path string, line int, offset int64, kind string, err error) error {
	return &CorruptionError{Path: path, Line: line, Offset: offset, Kind: kind, Detail: err.Error(), cause: err}
}

func manifestPath(path string) string { return path + ".manifest" }

func ledgerManifestFor(data []byte, records []Evidence) ledgerManifest {
	result := ledgerManifest{Schema: SchemaVersion, Bytes: int64(len(data)), Lines: len(records), Digest: digestBytes(data)}
	if len(records) > 0 {
		result.LastID = records[len(records)-1].ID
		result.LastHash = records[len(records)-1].Hash
	}
	return result
}

func verifyManifest(path string, state ledgerState) error {
	manifestData, err := os.ReadFile(manifestPath(path))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read provenance manifest: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestData))
	decoder.DisallowUnknownFields()
	var manifest ledgerManifest
	if err := decoder.Decode(&manifest); err != nil {
		return corruption(path, 0, 0, "manifest-malformed", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return corruption(path, 0, 0, "manifest-malformed", fmt.Errorf("manifest must contain one JSON value"))
	}
	expected := ledgerManifestFor(state.bytes, state.records)
	if manifest != expected {
		return corruption(path, 0, 0, "ledger-mutation", fmt.Errorf("manifest does not match canonical ledger state"))
	}
	return nil
}

func writeManifest(path string, manifest ledgerManifest) error {
	directory := filepath.Dir(manifestPath(path))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create provenance manifest directory: %w", err)
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal provenance manifest: %w", err)
	}
	payload = append(payload, '\n')
	temporary, err := os.CreateTemp(directory, ".provenance-manifest-*")
	if err != nil {
		return fmt.Errorf("create provenance manifest: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := writeFull(temporary, payload); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write provenance manifest: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync provenance manifest: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close provenance manifest: %w", err)
	}
	if err := os.Rename(temporaryName, manifestPath(path)); err != nil {
		return fmt.Errorf("commit provenance manifest: %w", err)
	}
	return nil
}

func checkReadOptions(records []Evidence, options ReadOptions) error {
	expected := strings.ToLower(strings.TrimSpace(options.ExpectedSourceDigest))
	legacyExpected := strings.ToLower(strings.TrimSpace(options.ExpectedSourceHash))
	if expected != "" && legacyExpected != "" && expected != legacyExpected {
		return fmt.Errorf("expected source digest and source hash differ")
	}
	if expected == "" {
		expected = legacyExpected
	}
	if expected != "" {
		if _, err := normalizeDigest(expected, "expected source digest"); err != nil {
			return err
		}
	}
	for _, record := range records {
		if expected != "" && record.SourceDigest != expected {
			return &FreshnessError{ID: record.ID, Kind: "source-mismatch", Expected: expected, Actual: record.SourceDigest}
		}
		if options.RequireFresh && record.Freshness.ValidUntil != "" {
			now := options.Now
			if now.IsZero() {
				now = time.Now().UTC()
			}
			validUntil, _ := time.Parse(time.RFC3339Nano, record.Freshness.ValidUntil)
			if !now.Before(validUntil) {
				return &FreshnessError{ID: record.ID, Kind: "expired", Actual: validUntil.Format(time.RFC3339Nano)}
			}
		}
	}
	return checkVerifiedClaims(records, options.RequiredVerified)
}

func checkVerifiedClaims(records []Evidence, claims []VerifiedClaim) error {
	for _, claim := range claims {
		semanticID, err := normalizeIdentifier(claim.SemanticID, "verified claim semantic_id")
		if err != nil {
			return &ClaimError{SemanticID: claim.SemanticID, Kind: "invalid", Detail: err.Error()}
		}
		semanticDigest, err := normalizeDigest(claim.SemanticDigest, "verified claim semantic_digest")
		if err != nil {
			return &ClaimError{SemanticID: semanticID, Kind: "invalid", Detail: err.Error()}
		}
		graphDigest, err := normalizeDigest(claim.GraphDigest, "verified claim graph_digest")
		if err != nil {
			return &ClaimError{SemanticID: semanticID, Kind: "invalid", Detail: err.Error()}
		}
		seen := false
		matched := false
		for _, record := range records {
			if record.SemanticID != semanticID {
				continue
			}
			seen = true
			if record.Status != StatusVerified {
				continue
			}
			if record.SemanticDigest == semanticDigest && record.GraphDigest == graphDigest {
				matched = true
				break
			}
		}
		if !seen || !matched {
			return &ClaimError{SemanticID: semanticID, Kind: "status-or-digest", Detail: "no verified event has both requested digests"}
		}
	}
	return nil
}

// ReadAll is a convenience alias for callers that want the explicit ledger
// wording in an adapter.
func (s *Store) ReadAll(options ReadOptions) (Snapshot, error) { return s.Read(options) }
