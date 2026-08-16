package semantic

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidFact     = errors.New("invalid semantic fact")
	ErrUnknownRelation = errors.New("unknown semantic relation")
)

// Relation is a PROV-inspired directed edge. The subject and object direction
// follows the PROV-O spelling: an Entity wasGeneratedBy an Activity, while an
// Activity used an Entity.
type Relation string

// RelationKind is the descriptive name used by query/projection adapters.
type RelationKind = Relation

type Predicate = Relation

// RelationSpec is the closed semantic type signature for a PROV-inspired
// relation. Keeping the signatures in one registry prevents the accepted
// vocabulary and its legality rules from drifting apart.
type RelationSpec struct {
	Predicate   Relation
	SubjectKind Kind
	ObjectKind  Kind
}

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

var relationSpecs = [...]RelationSpec{
	{Predicate: Used, SubjectKind: Activity, ObjectKind: Entity},
	{Predicate: WasGeneratedBy, SubjectKind: Entity, ObjectKind: Activity},
	{Predicate: WasDerivedFrom, SubjectKind: Entity, ObjectKind: Entity},
	{Predicate: WasAssociatedWith, SubjectKind: Activity, ObjectKind: Agent},
}

// RelationSpecs returns the deterministic, typed PROV-inspired vocabulary.
// The returned slice is detached so callers cannot mutate the registry.
func RelationSpecs() []RelationSpec {
	return append([]RelationSpec(nil), relationSpecs[:]...)
}

func (r Relation) Spec() (RelationSpec, bool) {
	for _, spec := range relationSpecs {
		if spec.Predicate == r {
			return spec, true
		}
	}
	return RelationSpec{}, false
}

func (r Relation) String() string {
	return string(r)
}

func (r Relation) Valid() bool {
	_, ok := r.Spec()
	return ok
}

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

func NewCandidateFact(subject ID, predicate Relation, object ID, reason string) Fact {
	return Fact{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
		Status:    FactCandidate,
		Reason:    strings.TrimSpace(reason),
	}
}

func NewUsedFact(activity, entity ID) Fact {
	return NewFact(activity, Used, entity)
}

func NewWasGeneratedByFact(entity, activity ID) Fact {
	return NewFact(entity, WasGeneratedBy, activity)
}

func NewWasDerivedFromFact(entity, source ID) Fact {
	return NewFact(entity, WasDerivedFrom, source)
}

func NewWasAssociatedWithFact(activity, agent ID) Fact {
	return NewFact(activity, WasAssociatedWith, agent)
}

func (f Fact) Key() FactKey {
	return FactKey{Subject: f.Subject, Predicate: f.Predicate, Object: f.Object}
}

func (f Fact) Normalized() (Fact, error) {
	subject, err := ParseIdentity(f.Subject.String())
	if err != nil {
		return Fact{}, fmt.Errorf("%w: subject: %v", ErrInvalidFact, err)
	}
	object, err := ParseIdentity(f.Object.String())
	if err != nil {
		return Fact{}, fmt.Errorf("%w: object: %v", ErrInvalidFact, err)
	}
	if !f.Predicate.Valid() {
		return Fact{}, fmt.Errorf("%w: %v", ErrUnknownRelation, f.Predicate)
	}
	if f.Status == 0 {
		// The zero value is useful for callers building literals; unspecified
		// status is conservative only in the sense that it remains deterministic
		// and can never become a candidate accidentally.
		f.Status = FactDeterministic
	}
	if f.Status != FactDeterministic && f.Status != FactCandidate {
		return Fact{}, fmt.Errorf("%w: unknown status %d", ErrInvalidFact, f.Status)
	}
	span := f.Span.Normalized()
	if err := span.Validate(); err != nil {
		return Fact{}, fmt.Errorf("%w: span: %v", ErrInvalidFact, err)
	}

	f.Subject = subject
	f.Object = object
	f.Span = span
	f.Reason = strings.TrimSpace(f.Reason)
	return f, nil
}

func (f Fact) Validate() error {
	_, err := f.Normalized()
	return err
}

func (f Fact) WithSpan(span Span) Fact {
	f.Span = span
	return f
}

func (f Fact) WithReason(reason string) Fact {
	f.Reason = strings.TrimSpace(reason)
	return f
}

// ActivityContract is the compact semantic form from which deterministic PROV
// facts can be derived. It does not depend on any parser or AST type.
type ActivityContract struct {
	Activity ID
	Inputs   []ID
	Outputs  []ID
	Agents   []ID
	Span     Span
}
