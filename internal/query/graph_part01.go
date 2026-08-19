package query

// Graph is a small in-memory relation store. Its zero value is ready for use;
// New is provided for callers that prefer an explicit constructor.
type Graph struct {
	nodes         map[ID]Node
	deterministic map[FactKey]Fact
	candidates    map[FactKey]Fact
	binding       *projectionBinding
}
type projectionBinding struct {
	namespace        string
	semanticDigest   string
	sourceDigest     string
	evidenceDigest   string
	provenanceDigest string
	sourceStatus     string
	evidenceStatus   string
	provenanceStatus string
}

func New() *Graph {
	graph := &Graph{}
	graph.ensure()
	return graph
}
func (graph *Graph) ensure() {
	if graph.nodes == nil {
		graph.nodes = make(map[ID]Node)
	}
	if graph.deterministic == nil {
		graph.deterministic = make(map[FactKey]Fact)
	}
	if graph.candidates == nil {
		graph.candidates = make(map[FactKey]Fact)
	}
}
