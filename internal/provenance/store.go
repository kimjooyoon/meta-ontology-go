package provenance

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Store is an append-only canonical JSONL evidence store. A Store serializes
// operations made through the same instance; separate processes need an
// external coordination mechanism if they write the same path concurrently.
type Store struct {
	path string
	mu   sync.Mutex
}

// New returns a store backed by path. The file is created by the first append.
func New(path string) *Store { return &Store{path: path} }

// Path returns the configured JSONL path.
func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// Append validates and appends one or more unique evidence entities. The batch
// is sorted by stable ID before writing, while prior lines remain untouched.
func (s *Store) Append(records ...Evidence) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return fmt.Errorf("provenance store path is required")
	}
	if len(records) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.readUnlocked(ReadOptions{})
	if err != nil {
		return err
	}
	known := make(map[string]struct{}, len(current.Records)+len(records))
	for _, record := range current.Records {
		known[record.ID] = struct{}{}
	}
	batch := make([]Evidence, 0, len(records))
	for index, record := range records {
		normalized, err := prepareEvidence(record)
		if err != nil {
			return fmt.Errorf("evidence %d: %w", index, err)
		}
		if _, exists := known[normalized.ID]; exists {
			return fmt.Errorf("evidence ID %q already exists", normalized.ID)
		}
		known[normalized.ID] = struct{}{}
		batch = append(batch, normalized)
	}
	sortEvidence(batch)
	payload, err := canonicalBatch(batch)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create provenance directory: %w", err)
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open provenance store: %w", err)
	}
	defer file.Close()
	if err := writeFull(file, payload); err != nil {
		return fmt.Errorf("append provenance evidence: %w", err)
	}
	return nil
}

// Read validates every line and returns records in stable ID order.
func (s *Store) Read(options ReadOptions) (Snapshot, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return Snapshot{}, fmt.Errorf("provenance store path is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readUnlocked(options)
}

func (s *Store) readUnlocked(options ReadOptions) (Snapshot, error) {
	records, err := s.readRecords()
	if err != nil {
		return Snapshot{}, err
	}
	for _, record := range records {
		if err := checkFreshness(record, options); err != nil {
			return Snapshot{}, err
		}
	}
	sortEvidence(records)
	digest, err := snapshotDigest(records)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Records: records, Digest: digest}, nil
}

func (s *Store) readRecords() ([]Evidence, error) {
	file, err := os.Open(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open provenance store: %w", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	records := make([]Evidence, 0)
	known := make(map[string]int)
	var offset int64
	for lineNumber := 1; ; lineNumber++ {
		raw, readErr := reader.ReadBytes('\n')
		if len(raw) > 0 {
			record, err := parseLine(s.path, lineNumber, offset, raw)
			if err != nil {
				return nil, err
			}
			if previous, exists := known[record.ID]; exists {
				return nil, corruption(s.path, lineNumber, offset, "duplicate-id", fmt.Errorf("ID %q already appeared on line %d", record.ID, previous))
			}
			known[record.ID] = lineNumber
			records = append(records, record)
			offset += int64(len(raw))
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read provenance store: %w", readErr)
		}
	}
	return records, nil
}

func parseLine(path string, lineNumber int, offset int64, raw []byte) (Evidence, error) {
	if raw[len(raw)-1] != '\n' {
		return Evidence{}, corruption(path, lineNumber, offset, "missing-newline", fmt.Errorf("canonical JSONL records must end with LF"))
	}
	content := bytes.TrimSuffix(raw, []byte{'\n'})
	if len(content) == 0 {
		return Evidence{}, corruption(path, lineNumber, offset, "blank-line", fmt.Errorf("blank lines are not valid evidence records"))
	}
	evidence, err := decodeEvidence(content)
	if err != nil {
		return Evidence{}, lineDiagnostic(path, lineNumber, offset, err)
	}
	normalized, err := normalizeEvidence(evidence)
	if err != nil {
		return Evidence{}, corruption(path, lineNumber, offset, "invalid-record", err)
	}
	unsigned, err := marshalEvidence(normalized, false)
	if err != nil {
		return Evidence{}, corruption(path, lineNumber, offset, "invalid-record", err)
	}
	expectedHash := digestBytes(unsigned)
	if evidence.Hash != expectedHash {
		return Evidence{}, corruption(path, lineNumber, offset, "hash-mismatch", fmt.Errorf("expected %q, got %q", expectedHash, evidence.Hash))
	}
	normalized.Hash = expectedHash
	expectedLine, err := marshalEvidence(normalized, true)
	if err != nil {
		return Evidence{}, corruption(path, lineNumber, offset, "invalid-record", err)
	}
	if !bytes.Equal(content, expectedLine) {
		return Evidence{}, corruption(path, lineNumber, offset, "non-canonical", fmt.Errorf("line does not match canonical JSON encoding"))
	}
	return normalized, nil
}

func prepareEvidence(evidence Evidence) (Evidence, error) {
	normalized, err := normalizeEvidence(evidence)
	if err != nil {
		return Evidence{}, err
	}
	unsigned, err := marshalEvidence(normalized, false)
	if err != nil {
		return Evidence{}, err
	}
	expectedHash := digestBytes(unsigned)
	if normalized.Hash != "" && strings.ToLower(strings.TrimSpace(normalized.Hash)) != expectedHash {
		return Evidence{}, fmt.Errorf("hash does not match canonical evidence content")
	}
	normalized.Hash = expectedHash
	return normalized, nil
}

func canonicalBatch(records []Evidence) ([]byte, error) {
	var payload bytes.Buffer
	for _, record := range records {
		line, err := marshalEvidence(record, true)
		if err != nil {
			return nil, err
		}
		payload.Write(line)
		payload.WriteByte('\n')
	}
	return payload.Bytes(), nil
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

func sortEvidence(records []Evidence) {
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
}

func snapshotDigest(records []Evidence) (string, error) {
	payload, err := canonicalBatch(records)
	if err != nil {
		return "", err
	}
	return digestBytes(payload), nil
}

func checkFreshness(record Evidence, options ReadOptions) error {
	expected := strings.ToLower(strings.TrimSpace(options.ExpectedSourceHash))
	if expected != "" {
		if err := validateDigest(expected); err != nil {
			return fmt.Errorf("expected source hash: %w", err)
		}
		if expected != record.Freshness.SourceHash {
			return &FreshnessError{ID: record.ID, Kind: "source-mismatch", Expected: expected, Actual: record.Freshness.SourceHash}
		}
	}
	if !options.RequireFresh || record.Freshness.ValidUntil == "" {
		return nil
	}
	now := options.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	validUntil, _ := time.Parse(time.RFC3339Nano, record.Freshness.ValidUntil)
	if now.After(validUntil) {
		return &FreshnessError{ID: record.ID, Kind: "expired", Actual: validUntil.Format(time.RFC3339Nano)}
	}
	return nil
}

func lineDiagnostic(path string, lineNumber int, offset int64, err error) error {
	var detail *lineError
	if !asLineError(err, &detail) {
		return corruption(path, lineNumber, offset, "invalid-record", err)
	}
	return corruption(path, lineNumber, offset, detail.kind, detail.err)
}

func asLineError(err error, target **lineError) bool {
	for err != nil {
		if candidate, ok := err.(*lineError); ok {
			*target = candidate
			return true
		}
		unwrapped, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapped.Unwrap()
	}
	return false
}

func corruption(path string, line int, offset int64, kind string, err error) error {
	return &CorruptionError{Path: path, Line: line, Offset: offset, Kind: kind, Detail: err.Error()}
}
