package provenance

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (s *Store) appendRecords(records []Evidence) error {
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
			return &DuplicateError{ID: normalized.ID}
		}
		known[normalized.ID] = struct{}{}
		batch = append(batch, normalized)
	}
	owners := predecessorOwners(current.Records)
	for _, record := range batch {
		if err := checkPredecessorClaims(owners, record); err != nil {
			return err
		}
	}
	sortEvidence(batch)
	payload, err := canonicalBatch(batch)
	if err != nil {
		return err
	}
	return appendPayload(s.path, payload)
}

func appendPayload(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create provenance directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open provenance store: %w", err)
	}
	defer file.Close()
	if err := writeFull(file, payload); err != nil {
		return fmt.Errorf("append provenance evidence: %w", err)
	}
	return nil
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
