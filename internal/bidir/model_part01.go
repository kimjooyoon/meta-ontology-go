package bidir

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

// ID is the stable identity of a semantic node.
type ID string

// Kind is open-ended so ontology extensions do not require this package to
// know every display kind.
type Kind string

const (
	EntityKind   Kind = "Entity"
	ActivityKind Kind = "Activity"
	AgentKind    Kind = "Agent"
)

// Predicate identifies a directed semantic relation.
type Predicate string

const (
	PredicateUsed           Predicate = "prov:used"
	PredicateWasGeneratedBy Predicate = "prov:wasGeneratedBy"
	PredicateWasDerivedFrom Predicate = "prov:wasDerivedFrom"
	PredicateInvokes        Predicate = "gooo:invokes"
	PredicateRepresents     Predicate = "gooo:represents"
	PredicateSpecialization Predicate = "prov:specializationOf"
)

// Reference names a declaration from a parser-neutral document.
type Reference struct {
	ID        ID
	Name      string
	Namespace string
	Span      SourceSpan
}

// Declaration is the parser-neutral representation of a DSL declaration.
type Declaration struct {
	Kind       Kind
	ID         ID
	Name       string
	Fields     []Field
	Inputs     []Reference
	Outputs    []Reference
	Attributes map[string]string
	Span       SourceSpan
}

// Document is the generic source view consumed by Get and Put.
type Document struct {
	Package               string
	Namespace             string
	Declarations          []Declaration
	Policies              []semantic.Policy
	Relations             []Relation
	ImplicitActivityPorts bool
}

// Node is a semantic declaration. Display fields do not define identity.
type Node struct {
	ID         ID
	Kind       Kind
	Name       string
	Namespace  string
	Aliases    []string
	Fields     []Field
	Attributes map[string]string
	Span       SourceSpan
}
