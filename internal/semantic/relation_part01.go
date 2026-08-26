package semantic

import (
	"errors"
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
