package impactgraph

import (
	"fmt"
	"sort"
)

// Normalized returns a detached graph with canonical ordering.
func (graph Graph) Normalized() (Graph, error) {
	if err := validateGraphHeader(graph); err != nil {
		return Graph{}, err
	}
	nodes, byID, err := normalizeNodes(graph.Nodes)
	if err != nil {
		return Graph{}, err
	}
	edges, err := normalizeEdges(graph.Edges, byID)
	if err != nil {
		return Graph{}, err
	}
	return Graph{Version: graph.Version, SnapshotDigest: graph.SnapshotDigest,
		RegistryDigest: graph.RegistryDigest, PolicyDigest: graph.PolicyDigest,
		Nodes: nodes, Edges: edges}, nil
}
func validateGraphHeader(graph Graph) error {
	if graph.Version != SchemaVersion {
		return fmt.Errorf("%w: unsupported schema version %q", ErrInvalidGraph, graph.Version)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "snapshot digest", value: graph.SnapshotDigest},
		{name: "registry digest", value: graph.RegistryDigest},
		{name: "policy digest", value: graph.PolicyDigest},
	} {
		if !validDigest(field.value) {
			return fmt.Errorf("%w: %s must be lowercase SHA-256 hex", ErrInvalidGraph, field.name)
		}
	}
	if len(graph.Nodes) == 0 {
		return fmt.Errorf("%w: no nodes", ErrInvalidGraph)
	}
	return nil
}
func normalizeNodes(rawNodes []Node) ([]Node, map[string]NodeKind, error) {
	nodes := make([]Node, len(rawNodes))
	byID := make(map[string]NodeKind, len(rawNodes))
	for index, node := range rawNodes {
		if err := validateID(node.ID); err != nil {
			return nil, nil, fmt.Errorf("%w at node %d: %v", ErrInvalidNode, index, err)
		}
		if !validNodeKind(node.Kind) {
			return nil, nil, fmt.Errorf("%w at node %d: unknown kind %q", ErrInvalidNode, index, node.Kind)
		}
		if _, exists := byID[node.ID]; exists {
			return nil, nil, fmt.Errorf("%w: %q", ErrDuplicateNode, node.ID)
		}
		byID[node.ID] = node.Kind
		nodes[index] = node
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].ID != nodes[j].ID {
			return nodes[i].ID < nodes[j].ID
		}
		return nodes[i].Kind < nodes[j].Kind
	})
	return nodes, byID, nil
}
