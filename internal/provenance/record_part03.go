package provenance

import (
	"errors"
	"fmt"
)

// Snapshot is the physical, validated ledger view. Digest covers exact
// canonical JSONL bytes in append order, so reordering changes the digest.
type Snapshot struct {
	Records []Evidence
	Digest  string
}

var (
	// ErrConflict identifies an event ID or digest conflict.
	ErrConflict = errors.New("provenance event conflict")
	// ErrChainGap identifies a missing, reordered, or mismatched predecessor.
	ErrChainGap = errors.New("provenance chain gap")
	// ErrClaimNotVerified identifies candidate/deferred evidence used as proof.
	ErrClaimNotVerified = errors.New("provenance claim is not verified")
	// ErrStaleSource identifies evidence bound to a different source snapshot.
	ErrStaleSource = errors.New("provenance source is stale")
)

// ConflictError reports a differing canonical event with an existing ID.
type ConflictError struct {
	ID     string
	Field  string
	Detail string
}

func (e *ConflictError) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("event %q conflicts with existing canonical bytes: %s", e.ID, e.Detail)
	}
	return fmt.Sprintf("event %q conflicts in %s: %s", e.ID, e.Field, e.Detail)
}
func (e *ConflictError) Unwrap() error { return ErrConflict }

// ChainError reports a predecessor or sequence violation.
type ChainError struct {
	ID       string
	Expected string
	Actual   string
	Detail   string
}

func (e *ChainError) Error() string {
	return fmt.Sprintf("event %q has a provenance chain gap: %s (expected %q, got %q)", e.ID, e.Detail, e.Expected, e.Actual)
}
func (e *ChainError) Unwrap() error { return ErrChainGap }

// ClaimError reports why a verification claim cannot be satisfied.
type ClaimError struct {
	SemanticID string
	Kind       string
	Detail     string
}

func (e *ClaimError) Error() string {
	return fmt.Sprintf("semantic claim %q is not verified (%s): %s", e.SemanticID, e.Kind, e.Detail)
}
func (e *ClaimError) Unwrap() error { return ErrClaimNotVerified }
