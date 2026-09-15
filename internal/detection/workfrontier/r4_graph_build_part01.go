package workfrontier

import (
	"sort"
)

func buildR4Graph(input R4Input) (r4Graph, string) {
	graph, nodeSet, edges, reason := indexR4Paths(input.Paths)
	if reason != "" {
		return graph, reason
	}
	roots, reason := r4GraphRoots(input.RootObligationIDs, nodeSet)
	if reason != "" {
		return graph, reason
	}
	populateR4Reachability(&graph, input.Paths, roots)
	sealR4Graph(&graph, input, roots, edges)
	return graph, ""
}
func indexR4Paths(paths []RepairPath) (r4Graph, map[string]struct{}, []r4Edge, string) {
	graph := r4Graph{}
	nodeSet := make(map[string]struct{}, len(paths))
	pathIDs := make(map[string]struct{}, len(paths))
	edges := make([]r4Edge, 0, len(paths))
	for _, path := range paths {
		if path.StableID == "" || path.ObligationID == "" {
			return graph, nil, nil, R4ReasonMalformedGraph
		}
		if _, exists := pathIDs[path.StableID]; exists {
			return graph, nil, nil, R4ReasonMalformedGraph
		}
		pathIDs[path.StableID] = struct{}{}
		nodeSet[path.ObligationID] = struct{}{}
		seenPrerequisites := make(map[string]struct{}, len(path.PrerequisiteObligationIDs))
		for _, prerequisite := range path.PrerequisiteObligationIDs {
			if prerequisite == "" {
				return graph, nil, nil, R4ReasonMalformedGraph
			}
			if _, duplicate := seenPrerequisites[prerequisite]; duplicate {
				return graph, nil, nil, R4ReasonMalformedGraph
			}
			seenPrerequisites[prerequisite] = struct{}{}
			nodeSet[prerequisite] = struct{}{}
			edges = append(edges, r4Edge{From: prerequisite, To: path.ObligationID, PathID: path.StableID})
		}
	}
	graph.nodes = sortedKeys(nodeSet)
	sort.Slice(edges, func(i, j int) bool { return edgeKey(edges[i]) < edgeKey(edges[j]) })
	graph.edges = edges
	return graph, nodeSet, edges, ""
}
func r4GraphRoots(values []string, nodes map[string]struct{}) ([]string, string) {
	if len(values) == 0 {
		return nil, R4ReasonRequiredInputMissing
	}
	if len(sortedUnique(values)) != len(values) {
		return nil, R4ReasonMalformedGraph
	}
	roots := sortedCopy(values)
	for _, root := range roots {
		if root == "" {
			return nil, R4ReasonMalformedGraph
		}
		if _, ok := nodes[root]; !ok {
			return nil, R4ReasonMalformedGraph
		}
	}
	return roots, ""
}
