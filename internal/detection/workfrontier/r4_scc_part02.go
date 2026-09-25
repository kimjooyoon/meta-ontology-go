package workfrontier

import (
	"sort"
)

func r4ComponentRecords(members [][]string, edges []r4Edge) []r4Component {
	componentIndex := make(map[string]int)
	for index, component := range members {
		for _, member := range component {
			componentIndex[member] = index
		}
	}
	result := make([]r4Component, 0, len(members))
	for index, componentMembers := range members {
		internalEdges := make([]r4Edge, 0)
		for _, edge := range edges {
			if componentIndex[edge.From] == index && componentIndex[edge.To] == index {
				internalEdges = append(internalEdges, edge)
			}
		}
		sort.Slice(internalEdges, func(i, j int) bool { return edgeKey(internalEdges[i]) < edgeKey(internalEdges[j]) })
		cyclic := len(componentMembers) > 1
		if !cyclic {
			for _, edge := range internalEdges {
				if edge.From == edge.To {
					cyclic = true
					break
				}
			}
		}
		result = append(result, r4Component{
			Digest:  digestR4(r4SCCPayload{Members: componentMembers, Edges: internalEdges}),
			Members: componentMembers, Edges: internalEdges, Cyclic: cyclic,
		})
	}
	return result
}
