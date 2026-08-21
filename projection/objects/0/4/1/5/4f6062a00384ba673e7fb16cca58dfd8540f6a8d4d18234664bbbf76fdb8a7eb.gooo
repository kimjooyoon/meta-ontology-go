package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func graphNodes(nodes []semantic.Node) []graphNode {
	nodes = canonicalNodes(nodes)
	result := make([]graphNode, 0, len(nodes))
	for _, node := range nodes {
		aliases := append([]string(nil), node.Aliases...)
		sort.Strings(aliases)
		fields := make([]graphField, 0, len(node.Fields))
		for _, field := range node.Fields {
			fieldAliases := append([]string(nil), field.Aliases...)
			sort.Strings(fieldAliases)
			fields = append(fields, graphField{
				ID: string(field.ID), Parent: string(field.Parent), Name: field.Name, Aliases: fieldAliases,
				TypeRefID: string(field.TypeRef.ID), Presence: string(field.Presence), Cardinality: string(field.Cardinality),
				Source: graphSpan{File: field.Span.File, Start: graphPosition{Offset: field.Span.Start.Offset, Line: field.Span.Start.Line, Column: field.Span.Start.Column}, End: graphPosition{Offset: field.Span.End.Offset, Line: field.Span.End.Line, Column: field.Span.End.Column}},
			})
		}
		result = append(result, graphNode{
			ID: string(node.ID), Kind: node.Kind.String(), Namespace: node.Namespace.String(),
			Name: node.Name, Aliases: aliases, Fields: fields,
		})
	}
	return result
}
