package cycles

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
