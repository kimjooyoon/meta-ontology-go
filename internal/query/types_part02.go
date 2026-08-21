package query

import (
	"fmt"
	"strings"
)

const (
	Used              Relation = "used"
	WasGeneratedBy    Relation = "wasGeneratedBy"
	WasDerivedFrom    Relation = "wasDerivedFrom"
	WasAssociatedWith Relation = "wasAssociatedWith"

	RelationUsed              = Used
	RelationWasGeneratedBy    = WasGeneratedBy
	RelationWasDerivedFrom    = WasDerivedFrom
	RelationWasAssociatedWith = WasAssociatedWith
)

// PROV-prefixed spellings are accepted at serialization/query boundaries and
// normalize to the local names used by the semantic IR.
const (
	PROVUsed              Relation = "prov:used"
	PROVWasGeneratedBy    Relation = "prov:wasGeneratedBy"
	PROVWasDerivedFrom    Relation = "prov:wasDerivedFrom"
	PROVWasAssociatedWith Relation = "prov:wasAssociatedWith"
)

// ParseRelation accepts canonical PROV names and their compact spellings.
func ParseRelation(raw Relation) (Relation, error) {
	switch strings.TrimSpace(string(raw)) {
	case string(Used), string(PROVUsed):
		return Used, nil
	case string(WasGeneratedBy), string(PROVWasGeneratedBy):
		return WasGeneratedBy, nil
	case string(WasDerivedFrom), string(PROVWasDerivedFrom):
		return WasDerivedFrom, nil
	case string(WasAssociatedWith), string(PROVWasAssociatedWith):
		return WasAssociatedWith, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidRelation, raw)
	}
}
func (relation Relation) String() string { return string(relation) }
func (relation Relation) Valid() bool {
	_, err := ParseRelation(relation)
	return err == nil
}

// FactStatus separates authoritative facts from observations that still need
// an assertion or review.
type FactStatus uint8

const (
	FactDeterministic FactStatus = iota + 1
	FactCandidate

	Deterministic = FactDeterministic
	Candidate     = FactCandidate
)

func (status FactStatus) String() string {
	switch status {
	case FactDeterministic:
		return "deterministic"
	case FactCandidate:
		return "candidate"
	default:
		return "unknown"
	}
}
