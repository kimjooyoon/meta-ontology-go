package provenance

import (
	"bytes"
	"fmt"
)

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
	combined := append(append([]byte(nil), state.bytes...), payload...)
	allRecords := append(append([]Evidence(nil), state.records...), pending...)
	if err := writeManifest(s.path, preparedManifestFor(state.bytes, state.records, combined, allRecords), false); err != nil {
		return fmt.Errorf("prepare provenance append: %w", err)
	}
	if err := appendPayload(s.path, payload); err != nil {
		return err
	}
	return writeManifest(s.path, ledgerManifestFor(combined, allRecords), false)
}
