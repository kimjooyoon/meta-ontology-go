package impactgraph

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

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

func normalizeEdges(rawEdges []Edge, byID map[string]NodeKind) ([]Edge, error) {
	edges := make([]Edge, len(rawEdges))
	seen := make(map[string]struct{}, len(rawEdges))
	for index, raw := range rawEdges {
		edge, err := raw.normalized()
		if err != nil {
			return nil, fmt.Errorf("%w at edge %d: %v", ErrInvalidEdge, index, err)
		}
		fromKind, fromOK := byID[edge.From]
		toKind, toOK := byID[edge.To]
		if !fromOK || !toOK {
			return nil, fmt.Errorf("%w at edge %d: endpoint is not registered", ErrInvalidEdge, index)
		}
		if !IsLegalEdge(edge.Kind, fromKind, toKind) {
			return nil, fmt.Errorf("%w at edge %d: %s %s->%s", ErrInvalidEdge, index, edge.Kind, fromKind, toKind)
		}
		key := edge.From + "\x00" + string(edge.Kind) + "\x00" + edge.To
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateEdge, key)
		}
		seen[key] = struct{}{}
		edges[index] = edge
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		return edges[i].Kind < edges[j].Kind
	})
	return edges, nil
}

func (edge Edge) normalized() (Edge, error) {
	from, err := edgeAlias(edge.From, edge.Source, edge.Subject, "from")
	if err != nil {
		return Edge{}, err
	}
	to, err := edgeAlias(edge.To, edge.Target, edge.Object, "to")
	if err != nil {
		return Edge{}, err
	}
	if err := validateID(from); err != nil {
		return Edge{}, fmt.Errorf("from: %v", err)
	}
	if err := validateID(to); err != nil {
		return Edge{}, fmt.Errorf("to: %v", err)
	}
	if _, known := endpointRules[edge.Kind]; !known {
		return Edge{}, fmt.Errorf("unknown kind %q", edge.Kind)
	}
	return Edge{From: from, To: to, Kind: edge.Kind}, nil
}

func edgeAlias(primary, first, second, field string) (string, error) {
	values := []string{primary, first, second}
	chosen := ""
	for _, value := range values {
		if value == "" {
			continue
		}
		if chosen != "" && chosen != value {
			return "", fmt.Errorf("conflicting %s aliases", field)
		}
		chosen = value
	}
	return chosen, nil
}

func validNodeKind(kind NodeKind) bool {
	switch kind {
	case NodeKindSource, NodeKindSemantic, NodeKindGoSymbol, NodeKindGoPackage,
		NodeKindGeneratedRegion, NodeKindObligation, NodeKindPressure:
		return true
	default:
		return false
	}
}

func validateID(value string) error {
	if value == "" || strings.TrimSpace(value) != value || strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return fmt.Errorf("stable ID must be non-empty and contain no whitespace or control characters")
	}
	return nil
}

func validDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
