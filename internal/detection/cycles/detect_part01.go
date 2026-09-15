package cycles

type nodeTable struct {
	nodes       map[ID]Node
	orderedIDs  []ID
	diagnostics Diagnostics
}
type normalizedEdge struct {
	subject   ID
	predicate Relation
	object    ID
	span      Span
	known     bool
}

// Detect reports every supported structural problem in graph. Diagnostics
// are sorted by a stable key, so equivalent graphs produce equivalent output
// regardless of declaration or relation insertion order.
func Detect(graph Graph) Diagnostics {
	table := indexNodes(graph.Nodes)
	edges, edgeDiagnostics := indexEdges(graph.edges(), table.nodes)
	result := append(table.diagnostics, edgeDiagnostics...)
	result = append(result, detectCycles(table.nodes, edges)...)
	sortDiagnostics(result)
	return result
}

// Analyze is a descriptive alias for Detect.
func Analyze(graph Graph) Diagnostics {
	return Detect(graph)
}

// DetectCycles is a descriptive alias for Detect. It retains the package's
// historical name while still returning all graph diagnostics.
func DetectCycles(graph Graph) Diagnostics {
	return Detect(graph)
}

// Validate is a descriptive alias for Detect.
func Validate(graph Graph) Diagnostics {
	return Detect(graph)
}

// Diagnostics returns the diagnostics for graph.
func (g Graph) Diagnostics() Diagnostics {
	return Detect(g)
}

// Validate returns the diagnostics for graph.
func (g Graph) Validate() Diagnostics {
	return Detect(g)
}

// Check returns nil for a valid graph, or its deterministic diagnostics as an
// error when one or more invariants fail.
func Check(graph Graph) error {
	diagnostics := Detect(graph)
	if len(diagnostics) == 0 {
		return nil
	}
	return diagnostics
}
