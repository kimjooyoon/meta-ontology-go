package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"sort"
)

func reachableCoverage(graph impactgraph.Graph, registry Registry, roots []string) ([]string, []string) {
	byID := make(map[string]impactgraph.NodeKind, len(graph.Nodes))
	adjacency := make(map[string][]string, len(graph.Nodes))
	for _, node := range graph.Nodes {
		byID[node.ID] = node.Kind
	}
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	registered := make(map[string]struct{}, len(registry.Obligations))
	for _, binding := range registry.Obligations {
		registered[binding.ID] = struct{}{}
	}
	requiredSet := make(map[string]struct{})
	uncovered := make([]string, 0)
	for _, root := range roots {
		reached := reachableFromCoverageRoot(root, adjacency, byID, registered)
		if len(reached) == 0 {
			uncovered = append(uncovered, root)
		}
		for _, obligation := range reached {
			requiredSet[obligation] = struct{}{}
		}
	}
	required := make([]string, 0, len(requiredSet))
	for obligation := range requiredSet {
		required = append(required, obligation)
	}
	sort.Strings(required)
	sort.Strings(uncovered)
	return required, uncovered
}
func reachableFromCoverageRoot(root string, adjacency map[string][]string, byID map[string]impactgraph.NodeKind, registered map[string]struct{}) []string {
	queue := []string{root}
	visited := map[string]struct{}{}
	reached := map[string]struct{}{}
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if _, seen := visited[current]; seen {
			continue
		}
		visited[current] = struct{}{}
		if byID[current] == impactgraph.NodeKindObligation {
			if _, ok := registered[current]; ok {
				reached[current] = struct{}{}
			}
		}
		queue = append(queue, adjacency[current]...)
	}
	result := make([]string, 0, len(reached))
	for obligation := range reached {
		result = append(result, obligation)
	}
	sort.Strings(result)
	return result
}
