package cycles

import (
	"sort"
)

func detectCycles(nodes map[ID]Node, edges []normalizedEdge) Diagnostics {
	adjacency := make(map[ID][]ID, len(nodes))
	for _, edge := range edges {
		if !edge.known {
			continue
		}
		adjacency[edge.subject] = append(adjacency[edge.subject], edge.object)
	}
	for id := range adjacency {
		adjacency[id] = uniqueSorted(adjacency[id])
	}
	ids := make([]ID, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	state := tarjanState{adjacency: adjacency, index: make(map[ID]int), low: make(map[ID]int), onStack: make(map[ID]bool)}
	for _, id := range ids {
		if _, visited := state.index[id]; !visited {
			state.visit(id)
		}
	}
	sort.Slice(state.components, func(i, j int) bool { return state.components[i][0] < state.components[j][0] })
	result := make(Diagnostics, 0)
	for _, component := range state.components {
		if !isCyclic(component, adjacency) {
			continue
		}
		path := cyclePath(component, adjacency)
		result = append(result, Diagnostic{
			Code: CycleDetected, NodeID: component[0], Cycle: path,
			Message: "directed cycle: " + joinIDs(path),
		})
	}
	return result
}

type tarjanState struct {
	adjacency  map[ID][]ID
	index      map[ID]int
	low        map[ID]int
	stack      []ID
	onStack    map[ID]bool
	next       int
	components [][]ID
}
