package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func resolveEndpoint(graph semantic.Graph, identity Identity) (semantic.ID, semantic.Node, error) {
	id, err := semantic.ParseIdentity(identity.ID)
	if err != nil {
		return "", semantic.Node{}, adapterError(AdapterUnknownEndpoint, "", identity.ID, err.Error())
	}
	node, ok := graph.Node(id)
	if !ok {
		return "", semantic.Node{}, adapterError(AdapterUnknownEndpoint, "", id.String(), "endpoint is not declared in the semantic graph")
	}
	return id, node, nil
}
func nodeKind(graph semantic.Graph, id semantic.ID) semantic.Kind {
	node, _ := graph.Node(id)
	return node.Kind
}
func semanticSpan(span Span) semantic.Span {
	return semantic.Span{
		File:  span.Filename,
		Start: semantic.Position{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:   semantic.Position{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}
