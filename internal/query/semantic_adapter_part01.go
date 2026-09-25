package query

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// FromSemanticIR builds a read-only query projection from a validated semantic
// IR. The semantic package remains authoritative; Graph is only a detached
// query view. Invalid IR, including unknown endpoints or invalid relation
// kinds, is rejected before any projection fact is added.
func FromSemanticIR(ir semantic.IR) (*Graph, error) {
	if err := ir.Validate(); err != nil {
		return nil, fmt.Errorf("semantic IR is not queryable: %w", err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		return nil, fmt.Errorf("semantic IR cannot be normalized for query: %w", err)
	}
	graph := New()
	evidenceStatus := "known_empty"
	if len(normalized.Evidence()) > 0 {
		evidenceStatus = "available"
	}
	graph.binding = &projectionBinding{
		namespace:      normalized.Namespace.String(),
		semanticDigest: normalized.StableHash(),
		sourceStatus:   "unavailable",
		evidenceDigest: normalized.EvidenceHash(),
		evidenceStatus: evidenceStatus,

		provenanceStatus: "unknown",
	}
	for _, node := range normalized.Graph.Nodes() {
		aliases := append([]string(nil), node.Aliases...)
		if err := graph.AddNode(Node{
			ID: ID(node.ID.String()), Kind: nodeKind(node.Kind),
			Namespace: node.Namespace.String(), Name: node.Name, Aliases: aliases,
		}); err != nil {
			return nil, fmt.Errorf("project semantic node: %w", err)
		}
	}
	for _, fact := range normalized.Graph.AllFacts() {
		projected, err := projectSemanticFact(fact)
		if err != nil {
			return nil, err
		}
		if err := graph.Add(projected); err != nil {
			return nil, fmt.Errorf("project semantic fact: %w", err)
		}
	}
	return graph, nil
}
func nodeKind(kind semantic.Kind) NodeKind {
	switch kind {
	case semantic.Entity:
		return EntityNodeKind
	case semantic.Activity:
		return ActivityNodeKind
	case semantic.Agent:
		return AgentNodeKind
	default:
		return UnknownNodeKind
	}
}
