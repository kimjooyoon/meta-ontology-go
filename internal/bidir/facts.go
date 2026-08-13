package bidir

import (
	"fmt"
	"sort"
)

// FactLayer records how strongly a fact is supported.
type FactLayer uint8

const (
	SyntacticFact FactLayer = iota + 1
	DeterministicFact
	CandidateFact
)

func (l FactLayer) String() string {
	switch l {
	case SyntacticFact:
		return "syntactic"
	case DeterministicFact:
		return "deterministic"
	case CandidateFact:
		return "candidate"
	default:
		return fmt.Sprintf("layer(%d)", l)
	}
}

// Fact is the adapter format for parser and Go analyzer observations.
type Fact struct {
	Layer     FactLayer
	Subject   ID
	Predicate Predicate
	Object    ID
	// EvidenceID is an adapter-supplied provenance identifier. It is kept out
	// of FactKey and semantic canonicalization so same-edge observations can
	// retain distinct evidence records without changing fact identity.
	EvidenceID  string
	SubjectKind Kind
	ObjectKind  Kind
	Attributes  map[string]string
	Source      SourceSpan
	Reason      string
}

// NewFact creates a fact without source metadata.
func NewFact(layer FactLayer, subject ID, predicate Predicate, object ID) Fact {
	return Fact{Layer: layer, Subject: subject, Predicate: predicate, Object: object}
}

// NewSourcedFact creates a source-backed fact.
func NewSourcedFact(layer FactLayer, subject ID, predicate Predicate, object ID, source SourceSpan) Fact {
	fact := NewFact(layer, subject, predicate, object)
	fact.Source = source
	return fact
}

// FactKey identifies a fact including its confidence layer.
type FactKey struct {
	Layer     FactLayer
	Subject   ID
	Predicate Predicate
	Object    ID
}

// Key returns layer-sensitive identity.
func (f Fact) Key() FactKey {
	return FactKey{Layer: f.Layer, Subject: f.Subject, Predicate: f.Predicate, Object: f.Object}
}

// SemanticKey returns layer-independent edge identity.
func (f Fact) SemanticKey() string {
	return relationKey(f.Predicate, f.Subject, f.Object)
}

func (f Fact) normalized() Fact {
	f.Attributes = cloneStringMap(f.Attributes)
	return f
}

// FactSet is a deterministic set-like collection.
type FactSet []Fact

// Normalized returns a sorted, deduplicated copy.
func (s FactSet) Normalized() FactSet {
	copySet := make(FactSet, len(s))
	for index, fact := range s {
		copySet[index] = fact.normalized()
	}
	sort.SliceStable(copySet, func(i, j int) bool { return factLess(copySet[i], copySet[j]) })
	result := make(FactSet, 0, len(copySet))
	seen := make(map[FactKey]struct{}, len(copySet))
	for _, fact := range copySet {
		if _, exists := seen[fact.Key()]; !exists {
			seen[fact.Key()] = struct{}{}
			result = append(result, fact)
		}
	}
	return result
}

// Normalize sorts and deduplicates the set in place.
func (s *FactSet) Normalize() {
	if s != nil {
		*s = s.Normalized()
	}
}

// ByLayer returns the deterministic subset with the requested layer.
func (s FactSet) ByLayer(layer FactLayer) FactSet {
	var result FactSet
	for _, fact := range s {
		if fact.Layer == layer {
			result = append(result, fact)
		}
	}
	return result.Normalized()
}

// Contains reports whether a fact with the same key exists.
func (s FactSet) Contains(candidate Fact) bool {
	key := candidate.Key()
	for _, fact := range s {
		if fact.Key() == key {
			return true
		}
	}
	return false
}

func (s FactSet) withoutKey(key FactKey) FactSet {
	result := make(FactSet, 0, len(s))
	for _, fact := range s {
		if fact.Key() != key {
			result = append(result, fact)
		}
	}
	return result
}

// withoutSemanticKey removes only candidate observations for an authoritative
// triple. Candidate identity includes its layer, but shadowing is deliberately
// layer-independent: a deterministic fact must not leave a stale candidate in
// the model while the raw/evidence boundary retains the observation.
func (s FactSet) withoutSemanticKey(key string) FactSet {
	result := make(FactSet, 0, len(s))
	for _, fact := range s {
		if fact.Layer == CandidateFact && fact.SemanticKey() == key {
			continue
		}
		result = append(result, fact)
	}
	return result
}

func factLess(left, right Fact) bool {
	if left.Layer != right.Layer {
		return left.Layer < right.Layer
	}
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	if left.Object != right.Object {
		return left.Object < right.Object
	}
	if left.Source.File != right.Source.File {
		return left.Source.File < right.Source.File
	}
	if left.Source.Start != right.Source.Start {
		return left.Source.Start < right.Source.Start
	}
	if left.Source.End != right.Source.End {
		return left.Source.End < right.Source.End
	}
	if left.Source.StartLine != right.Source.StartLine {
		return left.Source.StartLine < right.Source.StartLine
	}
	if left.Source.StartColumn != right.Source.StartColumn {
		return left.Source.StartColumn < right.Source.StartColumn
	}
	if left.Source.EndLine != right.Source.EndLine {
		return left.Source.EndLine < right.Source.EndLine
	}
	if left.Source.EndColumn != right.Source.EndColumn {
		return left.Source.EndColumn < right.Source.EndColumn
	}
	if left.SubjectKind != right.SubjectKind {
		return left.SubjectKind < right.SubjectKind
	}
	if left.ObjectKind != right.ObjectKind {
		return left.ObjectKind < right.ObjectKind
	}
	if left.Reason != right.Reason {
		return left.Reason < right.Reason
	}
	return attributesLess(left.Attributes, right.Attributes)
}

func attributesLess(left, right map[string]string) bool {
	leftKeys, rightKeys := mapKeys(left), mapKeys(right)
	for index := 0; index < len(leftKeys) && index < len(rightKeys); index++ {
		if leftKeys[index] != rightKeys[index] {
			return leftKeys[index] < rightKeys[index]
		}
		if left[leftKeys[index]] != right[rightKeys[index]] {
			return left[leftKeys[index]] < right[rightKeys[index]]
		}
	}
	return len(leftKeys) < len(rightKeys)
}

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

func (e *ReconcileError) Error() string {
	if len(e.Conflicts) == 0 {
		return "bidir reconciliation failed"
	}
	return fmt.Sprintf("bidir reconciliation rejected %d fact(s): %s", len(e.Conflicts), e.Conflicts[0].Message)
}

// ReconcileResult contains accepted layers, semantic delta, locality, and a
// detached non-authoritative raw observation boundary.
type ReconcileResult struct {
	Model          Model
	Delta          Delta
	Locality       Locality
	RawObservation RawFactObservation
	Accepted       FactSet
	Syntactic      FactSet
	Candidates     FactSet
	Conflicts      []Conflict
}
