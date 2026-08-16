package workfrontier

import "sort"

func deriveR4Components(nodes []string, edges []r4Edge) []r4Component {
	members := tarjanR4(nodes, edges)
	sort.Slice(members, func(i, j int) bool { return members[i][0] < members[j][0] })
	return r4ComponentRecords(members, edges)
}

func tarjanR4(nodes []string, edges []r4Edge) [][]string {
	adjacency := make(map[string][]string, len(nodes))
	for _, node := range nodes {
		adjacency[node] = nil
	}
	for _, edge := range edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	for node := range adjacency {
		adjacency[node] = sortedUnique(adjacency[node])
	}
	index := 0
	indices := make(map[string]int, len(nodes))
	lowlink := make(map[string]int, len(nodes))
	onStack := make(map[string]bool, len(nodes))
	stack := make([]string, 0, len(nodes))
	components := make([][]string, 0)
	var visit func(string)
	visit = func(node string) {
		indices[node], lowlink[node] = index, index
		index++
		stack = append(stack, node)
		onStack[node] = true
		for _, next := range adjacency[node] {
			if _, seen := indices[next]; !seen {
				visit(next)
				if lowlink[next] < lowlink[node] {
					lowlink[node] = lowlink[next]
				}
			} else if onStack[next] && indices[next] < lowlink[node] {
				lowlink[node] = indices[next]
			}
		}
		if lowlink[node] != indices[node] {
			return
		}
		component := make([]string, 0)
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			component = append(component, last)
			if last == node {
				break
			}
		}
		sort.Strings(component)
		components = append(components, component)
	}
	for _, node := range nodes {
		if _, seen := indices[node]; !seen {
			visit(node)
		}
	}
	return components
}

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
