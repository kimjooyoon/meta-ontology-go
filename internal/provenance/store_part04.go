package provenance

import (
	"bytes"
	"fmt"
)

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

// ReadAll is a convenience alias for adapters that prefer explicit wording.
func (s *Store) ReadAll(options ReadOptions) (Snapshot, error) { return s.Read(options) }

func validateDuplicateEvidence(normalized, existing Evidence, index int) error {
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
	return nil
}
