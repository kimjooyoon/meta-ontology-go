package semantic

import (
	"fmt"
)

// ValidateKinds checks the PROV-inspired type signature of a relation.
func (r Relation) ValidateKinds(subject, object Kind) error {
	spec, ok := r.Spec()
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownRelation, r)
	}
	if subject == spec.SubjectKind && object == spec.ObjectKind {
		return nil
	}
	return fmt.Errorf("%w: %s cannot connect %s to %s", ErrInvalidFact, r, subject, object)
}

// FactStatus distinguishes facts that the compiler can derive deterministically
// from facts that require an explicit assertion or later review.
type FactStatus uint8

const (
	FactDeterministic FactStatus = iota + 1
	FactCandidate

	Deterministic = FactDeterministic
	Candidate     = FactCandidate
)

func (s FactStatus) String() string {
	switch s {
	case FactDeterministic:
		return "deterministic"
	case FactCandidate:
		return "candidate"
	default:
		return "unknown"
	}
}

// FactKey identifies a relation independently of source evidence and status.
// A deterministic fact shadows a candidate with the same triple.
type FactKey struct {
	Subject   ID
	Predicate Relation
	Object    ID
}

// Fact is a directed semantic relation plus optional source context. Source
// spans and candidate reasons do not alter the identity of a triple.
type Fact struct {
	Subject   ID
	Predicate Relation
	Object    ID
	Status    FactStatus
	Span      Span
	Reason    string
}

func NewFact(subject ID, predicate Relation, object ID) Fact {
	return Fact{Subject: subject, Predicate: predicate, Object: object, Status: FactDeterministic}
}
func NewDeterministicFact(subject ID, predicate Relation, object ID) Fact {
	return NewFact(subject, predicate, object)
}
