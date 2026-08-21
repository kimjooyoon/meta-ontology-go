package cycles

// ID is a stable semantic identifier. It is deliberately an alias for string
// so callers can use IDs from another package without conversion ceremony.
type ID = string

// StableID is a descriptive alias for callers that use the full vocabulary.
type StableID = ID

// Kind is the PROV-inspired role of a graph node.
type Kind string

const (
	Entity   Kind = "Entity"
	Activity Kind = "Activity"
	Agent    Kind = "Agent"

	EntityKind   = Entity
	ActivityKind = Activity
	AgentKind    = Agent
)

// String returns the role spelling used in diagnostics.
func (k Kind) String() string {
	return string(k)
}

// Node is a semantic declaration. Name and Aliases are presentation and
// lookup metadata; ID, Kind, and Namespace define the identity boundary.
type Node struct {
	ID        ID
	Kind      Kind
	Namespace string
	Name      string
	Aliases   []string
	Span      Span
}

// Relation is a directed PROV-inspired predicate.
type Relation string

// Predicate is a descriptive alias for Relation.
type Predicate = Relation

const (
	Used              Relation = "used"
	WasGeneratedBy    Relation = "wasGeneratedBy"
	WasDerivedFrom    Relation = "wasDerivedFrom"
	WasAssociatedWith Relation = "wasAssociatedWith"
	WasInformedBy     Relation = "wasInformedBy"
	WasAttributedTo   Relation = "wasAttributedTo"
	ActedOnBehalfOf   Relation = "actedOnBehalfOf"
	SpecializationOf  Relation = "specializationOf"
	AlternateOf       Relation = "alternateOf"

	RelationUsed              = Used
	RelationWasGeneratedBy    = WasGeneratedBy
	RelationWasDerivedFrom    = WasDerivedFrom
	RelationWasAssociatedWith = WasAssociatedWith
)

// String returns the predicate spelling used in diagnostics.
func (r Relation) String() string {
	return string(r)
}
