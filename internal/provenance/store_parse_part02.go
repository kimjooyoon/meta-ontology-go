package provenance

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

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
