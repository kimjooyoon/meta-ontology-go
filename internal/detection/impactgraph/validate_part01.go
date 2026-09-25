package impactgraph

type endpointPair struct {
	from NodeKind
	to   NodeKind
}

var endpointRules = map[EdgeKind][]endpointPair{
	EdgeKindDeclares:   {{from: NodeKindSource, to: NodeKindSemantic}},
	EdgeKindImplements: {{from: NodeKindGoSymbol, to: NodeKindSemantic}},
	EdgeKindProjectsTo: {
		{from: NodeKindSemantic, to: NodeKindGoSymbol},
		{from: NodeKindSemantic, to: NodeKindGoPackage},
		{from: NodeKindSemantic, to: NodeKindGeneratedRegion},
	},
	EdgeKindImportAffects: {{from: NodeKindGoPackage, to: NodeKindGoPackage}},
	EdgeKindAffects: {
		{from: NodeKindSource, to: NodeKindObligation},
		{from: NodeKindSource, to: NodeKindPressure},
		{from: NodeKindSemantic, to: NodeKindObligation},
		{from: NodeKindSemantic, to: NodeKindPressure},
		{from: NodeKindGoSymbol, to: NodeKindObligation},
		{from: NodeKindGoSymbol, to: NodeKindPressure},
		{from: NodeKindGoPackage, to: NodeKindObligation},
		{from: NodeKindGoPackage, to: NodeKindPressure},
		{from: NodeKindGeneratedRegion, to: NodeKindObligation},
		{from: NodeKindGeneratedRegion, to: NodeKindPressure},
	},
	EdgeKindVerifiedBy: {
		{from: NodeKindObligation, to: NodeKindGoSymbol},
		{from: NodeKindObligation, to: NodeKindGoPackage},
	},
	EdgeKindMeasuredBy: {{from: NodeKindPressure, to: NodeKindObligation}},
}

// EndpointKinds returns the legal (from, to) pairs for an edge kind.
func EndpointKinds(kind EdgeKind) [][2]NodeKind {
	rules := endpointRules[kind]
	result := make([][2]NodeKind, 0, len(rules))
	for _, rule := range rules {
		result = append(result, [2]NodeKind{rule.from, rule.to})
	}
	return result
}

// IsLegalEdge reports whether kind permits the two endpoint kinds.
func IsLegalEdge(kind EdgeKind, from, to NodeKind) bool {
	for _, rule := range endpointRules[kind] {
		if rule.from == from && rule.to == to {
			return true
		}
	}
	return false
}

// Validate checks the graph without mutating it.
func (graph Graph) Validate() error {
	_, err := graph.Normalized()
	return err
}
