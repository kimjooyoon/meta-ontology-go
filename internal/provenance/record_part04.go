package provenance

import (
	"fmt"
)

// CorruptionError identifies a malformed or integrity-invalid ledger line or
// manifest. Byte offsets are offsets in the JSONL file.
type CorruptionError struct {
	Path   string
	Line   int
	Offset int64
	Kind   string
	Detail string
	cause  error
}

func (e *CorruptionError) Error() string {
	return fmt.Sprintf("provenance corruption at %s:%d (byte %d, %s): %s", e.Path, e.Line, e.Offset, e.Kind, e.Detail)
}
func (e *CorruptionError) Unwrap() error { return e.cause }

// FreshnessError reports a valid event that does not match the requested
// source snapshot or freshness window.
type FreshnessError struct {
	ID       string
	Kind     string
	Expected string
	Actual   string
}

func (e *FreshnessError) Error() string {
	if e.Kind == "source-mismatch" {
		return fmt.Sprintf("event %q is stale: source digest %q, expected %q", e.ID, e.Actual, e.Expected)
	}
	return fmt.Sprintf("event %q is stale: %s", e.ID, e.Kind)
}
func (e *FreshnessError) Unwrap() error { return ErrStaleSource }
