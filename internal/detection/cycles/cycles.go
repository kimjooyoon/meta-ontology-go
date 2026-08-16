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
	start := component[0]
	members := make(map[ID]struct{}, len(component))
	for _, id := range component {
		members[id] = struct{}{}
	}

	// A reverse breadth-first search finds the shortest deterministic route
	// back to the component's canonical start. It visits each component edge
	// at most once, keeping path construction bounded even for dense SCCs.
	reverse := make(map[ID][]ID, len(component))
	for _, id := range component {
		for _, next := range adjacency[id] {
			if _, member := members[next]; member {
				reverse[next] = append(reverse[next], id)
			}
		}
	}
	for id := range reverse {
		reverse[id] = uniqueSorted(reverse[id])
	}
	queue := []ID{start}
	nextTowardStart := map[ID]ID{start: ""}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, predecessor := range reverse[current] {
			if _, seen := nextTowardStart[predecessor]; seen {
				continue
			}
			nextTowardStart[predecessor] = current
			queue = append(queue, predecessor)
		}
	}

	for _, next := range adjacency[start] {
		if _, member := members[next]; !member {
			continue
		}
		if next == start {
			return []ID{start, start}
		}
		if _, reachable := nextTowardStart[next]; !reachable {
			continue
		}
		path := []ID{start, next}
		for current := nextTowardStart[next]; current != start; current = nextTowardStart[current] {
			path = append(path, current)
		}
		return append(path, start)
	}
	return []ID{start, start}
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
