package cycles

import "sort"

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

func (s *tarjanState) visit(id ID) {
	s.index[id] = s.next
	s.low[id] = s.next
	s.next++
	s.stack = append(s.stack, id)
	s.onStack[id] = true
	for _, next := range s.adjacency[id] {
		if _, visited := s.index[next]; !visited {
			s.visit(next)
			if s.low[next] < s.low[id] {
				s.low[id] = s.low[next]
			}
			continue
		}
		if s.onStack[next] && s.index[next] < s.low[id] {
			s.low[id] = s.index[next]
		}
	}
	if s.low[id] != s.index[id] {
		return
	}
	component := make([]ID, 0)
	for {
		last := len(s.stack) - 1
		member := s.stack[last]
		s.stack = s.stack[:last]
		s.onStack[member] = false
		component = append(component, member)
		if member == id {
			break
		}
	}
	sort.Strings(component)
	s.components = append(s.components, component)
}

func isCyclic(component []ID, adjacency map[ID][]ID) bool {
	if len(component) > 1 {
		return true
	}
	for _, next := range adjacency[component[0]] {
		if next == component[0] {
			return true
		}
	}
	return false
}

func cyclePath(component []ID, adjacency map[ID][]ID) []ID {
	members := make(map[ID]struct{}, len(component))
	for _, id := range component {
		members[id] = struct{}{}
	}
	path := []ID{component[0]}
	visited := map[ID]bool{component[0]: true}
	if findCyclePath(component[0], component[0], members, adjacency, visited, &path) {
		return path
	}
	return append(path, component[0])
}

func findCyclePath(current, start ID, members map[ID]struct{}, adjacency map[ID][]ID, visited map[ID]bool, path *[]ID) bool {
	for _, next := range adjacency[current] {
		if next == start {
			*path = append(*path, start)
			return true
		}
		if _, member := members[next]; !member || visited[next] {
			continue
		}
		visited[next] = true
		*path = append(*path, next)
		if findCyclePath(next, start, members, adjacency, visited, path) {
			return true
		}
		*path = (*path)[:len(*path)-1]
		delete(visited, next)
	}
	return false
}

func uniqueSorted(values []ID) []ID {
	result := append([]ID(nil), values...)
	sort.Strings(result)
	write := 0
	for _, value := range result {
		if write == 0 || result[write-1] != value {
			result[write] = value
			write++
		}
	}
	return result[:write]
}

func joinIDs(ids []ID) string {
	result := ""
	for i, id := range ids {
		if i > 0 {
			result += " -> "
		}
		result += id
	}
	return result
}
