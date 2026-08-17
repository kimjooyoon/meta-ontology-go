package selectiveci

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
)

func buildGraph(input Input) (impactgraph.Graph, error) {
	registry := input.Registry
	nodes := append([]impactgraph.Node(nil), registry.Nodes...)
	byID := map[string]impactgraph.NodeKind{}
	for _, node := range nodes {
		byID[node.ID] = node.Kind
	}
	for _, file := range append(append([]SnapshotFile{}, input.Base.Files...), input.Head.Files...) {
		for _, id := range file.SemanticIDs {
			if err := addNode(byID, &nodes, impactgraph.Node{ID: id, Kind: impactgraph.NodeKindSemantic}); err != nil {
				return impactgraph.Graph{}, err
			}
		}
	}
	edges := make([]impactgraph.Edge, 0, len(registry.DependencyEdges)+len(registry.Obligations))
	for _, edge := range registry.DependencyEdges {
		edges = append(edges, impactgraph.Edge{From: edge.From, To: edge.To, Kind: edge.Kind})
	}
	for _, binding := range registry.Obligations {
		if err := addNode(byID, &nodes, impactgraph.Node{ID: binding.ID, Kind: impactgraph.NodeKindObligation}); err != nil {
			return impactgraph.Graph{}, err
		}
		if err := addNode(byID, &nodes, impactgraph.Node{ID: binding.Subject, Kind: impactgraph.NodeKindSemantic}); err != nil {
			return impactgraph.Graph{}, err
		}
		edges = append(edges, impactgraph.Edge{From: binding.Subject, To: binding.ID, Kind: impactgraph.EdgeKindAffects})
	}
	graph := impactgraph.Graph{Version: impactgraph.SchemaVersion, SnapshotDigest: input.Head.Digest, RegistryDigest: registry.Digest, PolicyDigest: registry.PolicyDigest, Nodes: nodes, Edges: edges}
	if err := graph.Validate(); err != nil {
		return impactgraph.Graph{}, failure(graphFailureReason(err.Error()), err.Error())
	}
	return graph, nil
}

func addNode(byID map[string]impactgraph.NodeKind, nodes *[]impactgraph.Node, node impactgraph.Node) error {
	if kind, exists := byID[node.ID]; exists {
		if kind != node.Kind {
			return failure(ReasonEvaluatorError, "stable ID has conflicting node kinds")
		}
		return nil
	}
	byID[node.ID] = node.Kind
	*nodes = append(*nodes, node)
	return nil
}

func graphFailureReason(message string) string {
	message = strings.ToLower(message)
	if strings.Contains(message, "duplicate") {
		return ReasonDuplicateID
	}
	if strings.Contains(message, "endpoint") || strings.Contains(message, "registered") {
		return ReasonDanglingReference
	}
	return ReasonEvaluatorError
}
