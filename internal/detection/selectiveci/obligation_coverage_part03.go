package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"strings"
)

func validateCoverageBindings(graph impactgraph.Graph, registry Registry) CoverageReason {
	nodes := make(map[string]impactgraph.NodeKind, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node.Kind
	}
	bindings := make(map[string]struct{}, len(registry.Obligations))
	for _, binding := range registry.Obligations {
		bindings[binding.ID] = struct{}{}
		if nodes[binding.ID] != impactgraph.NodeKindObligation {
			return CoverageReasonStaleGraph
		}
	}
	for _, node := range graph.Nodes {
		if node.Kind == impactgraph.NodeKindObligation {
			if _, registered := bindings[node.ID]; !registered {
				return CoverageReasonStaleRegistry
			}
		}
	}
	return ""
}
func coverageRoots(raw []string, graph impactgraph.Graph) ([]string, CoverageReason) {
	roots := sortedCopy(raw)
	byID := make(map[string]impactgraph.NodeKind, len(graph.Nodes))
	for _, node := range graph.Nodes {
		byID[node.ID] = node.Kind
	}
	for index, root := range roots {
		if index > 0 && root == roots[index-1] {
			return nil, CoverageReasonDuplicateRoot
		}
		if root == "" || strings.TrimSpace(root) != root {
			return nil, CoverageReasonUnknownRoot
		}
		if byID[root] != impactgraph.NodeKindSemantic {
			return nil, CoverageReasonUnknownRoot
		}
	}
	return roots, ""
}
