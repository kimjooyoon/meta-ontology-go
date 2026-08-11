// Package cycles validates the structural invariants of a PROV-inspired graph.
//
// The package owns a small adapter-friendly graph model instead of importing
// the semantic kernel. Callers can populate it from a DSL IR, a Go analyzer,
// or a serialized graph without creating a dependency in either direction.
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

// Edge is a directed semantic relation. Relation is accepted as an alias for
// Predicate when adapting data that uses that field name.
type Edge struct {
	Subject   ID
	Predicate Relation
	Relation  Relation
	Object    ID
	Span      Span
}

// Fact is an alias for graph data sources that call edges facts.
type Fact = Edge

// Graph is an intentionally open input model. Edges and Relations are both
// accepted to make adapters from fact-oriented and relation-oriented models
// straightforward; callers should populate only one of them.
type Graph struct {
	Nodes     []Node
	Edges     []Edge
	Relations []Edge
}

// NewGraph returns an empty graph ready for AddNode and AddEdge.
func NewGraph() Graph {
	return Graph{}
}

// AddNode appends a declaration. Validation is deferred to Detect so all
// independent problems can be reported in one deterministic result.
func (g *Graph) AddNode(node Node) {
	if g == nil {
		return
	}
	g.Nodes = append(g.Nodes, node)
}

// AddEdge appends a directed relation.
func (g *Graph) AddEdge(edge Edge) {
	if g == nil {
		return
	}
	g.Edges = append(g.Edges, edge)
}

// AddRelation is an explicit synonym for AddEdge.
func (g *Graph) AddRelation(edge Edge) {
	g.AddEdge(edge)
}

// AddFact is an explicit synonym for AddEdge.
func (g *Graph) AddFact(fact Fact) {
	g.AddEdge(fact)
}

func (e Edge) predicate() Relation {
	if e.Predicate != "" {
		return e.Predicate
	}
	return e.Relation
}

func (g Graph) edges() []Edge {
	result := make([]Edge, 0, len(g.Edges)+len(g.Relations))
	result = append(result, g.Edges...)
	result = append(result, g.Relations...)
	return result
}

// Position identifies a source location. Zero values are valid and mean that
// the source location is unavailable.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span carries optional source provenance into diagnostics.
type Span struct {
	File  string
	Start Position
	End   Position
}
