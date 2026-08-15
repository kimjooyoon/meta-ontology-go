package provenance

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

func parseLedgerData(path string, data []byte) (ledgerState, error) {
	state := ledgerState{bytes: append([]byte(nil), data...), digest: digestBytes(data)}
	if len(data) == 0 {
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
		if err := validatePredecessor(state.records, record, path, lineNumber, offset); err != nil {
			return ledgerState{}, err
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
	return state, nil
}

func validatePredecessor(records []Evidence, record Evidence, path string, line int, offset int64) error {
	if len(records) == 0 {
		if record.Predecessor != nil {
			return corruption(path, line, offset, "chain-gap", fmt.Errorf("first record has a predecessor"))
		}
		return nil
	}
	previous := records[len(records)-1]
	expected := &DigestLink{ID: previous.ID, Digest: previous.Hash}
	if record.Predecessor == nil || record.Predecessor.ID != expected.ID || record.Predecessor.Digest != expected.Digest {
		return corruption(path, line, offset, "chain-gap", fmt.Errorf("predecessor does not identify line %d", line-1))
	}
	return nil
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
