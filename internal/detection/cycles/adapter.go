package cycles

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

// FromSemanticGraph adapts the semantic kernel's normalized graph into the
// detector's dependency-light input model. Candidates are included because
// structural direction and identity checks apply before promotion.
func FromSemanticGraph(source semantic.Graph) Graph {
	nodes := source.SortedNodes()
	result := Graph{
		Nodes: make([]Node, 0, len(nodes)),
		Edges: make([]Edge, 0, len(source.AllFacts())),
	}
	for _, node := range nodes {
		result.Nodes = append(result.Nodes, Node{
			ID: node.ID.String(), Kind: Kind(node.Kind.String()),
			Namespace: node.Namespace.String(), Name: node.Name,
			Aliases: append([]string(nil), node.Aliases...), Span: adaptSpan(node.Span),
		})
	}
	for _, fact := range source.AllFacts() {
		result.Edges = append(result.Edges, Edge{
			Subject: fact.Subject.String(), Predicate: Relation(fact.Predicate.String()),
			Object: fact.Object.String(), Span: adaptSpan(fact.Span),
		})
	}
	return result
}

// DetectSemanticGraph validates a semantic-kernel graph without changing it.
func DetectSemanticGraph(source semantic.Graph) Diagnostics {
	return Detect(FromSemanticGraph(source))
}

func adaptSpan(source semantic.Span) Span {
	return Span{
		File:  source.File,
		Start: Position{Offset: source.Start.Offset, Line: source.Start.Line, Column: source.Start.Column},
		End:   Position{Offset: source.End.Offset, Line: source.End.Line, Column: source.End.Column},
	}
}
