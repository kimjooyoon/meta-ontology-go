package bidir

import (
	"sort"
)

func mapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// FactDelta is an explicit add/remove update. Absence is not deletion.
type FactDelta struct {
	Added   FactSet
	Removed FactSet
}

// Normalized returns a deterministic fact delta.
func (d FactDelta) Normalized() FactDelta {
	return FactDelta{Added: d.Added.Normalized(), Removed: d.Removed.Normalized()}
}

// RawFactObservation preserves adapter observations before semantic FactSet
// normalization. It is detached, non-authoritative evidence: duplicate
// FactKeys remain distinct when their evidence records differ.
type RawFactObservation struct {
	Added        FactSet
	Removed      FactSet
	EvidenceHash string
}

// ReconcileOptions controls provenance strictness.
type ReconcileOptions struct {
	RequireSource bool
}

// DefaultReconcileOptions is the strict CI policy.
func DefaultReconcileOptions() ReconcileOptions { return ReconcileOptions{RequireSource: true} }

// PermissiveReconcileOptions allows trusted adapters to omit spans.
func PermissiveReconcileOptions() ReconcileOptions { return ReconcileOptions{} }

// ConflictKind identifies a rejected semantic update.
type ConflictKind string

const (
	ConflictMissingSource    ConflictKind = "missing-source"
	ConflictUnknownPredicate ConflictKind = "unknown-predicate"
	ConflictUnknownEndpoint  ConflictKind = "unknown-endpoint"
	ConflictKindMismatch     ConflictKind = "kind-mismatch"
	ConflictInvalidFact      ConflictKind = "invalid-fact"
)

// Conflict explains why one fact was rejected.
type Conflict struct {
	Kind    ConflictKind
	Fact    Fact
	Message string
}

// ReconcileError reports rejected facts while preserving transactionality.
type ReconcileError struct {
	Conflicts []Conflict
}
