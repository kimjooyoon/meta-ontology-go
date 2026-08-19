package cycles

import (
	"sort"
)

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
