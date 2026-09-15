package impactgraph

import (
	"fmt"
	"sort"
)

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
