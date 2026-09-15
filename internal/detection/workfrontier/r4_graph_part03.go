package workfrontier

func reachableFrom(roots []string, adjacency map[string][]string) map[string]struct{} {
	seen := make(map[string]struct{}, len(adjacency))
	stack := append([]string(nil), roots...)
	for len(stack) != 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, exists := seen[node]; exists {
			continue
		}
		seen[node] = struct{}{}
		for _, next := range adjacency[node] {
			stack = append(stack, next)
		}
	}
	return seen
}
func countCyclicR4Components(components []r4Component) int {
	count := 0
	for _, component := range components {
		if component.Cyclic {
			count++
		}
	}
	return count
}
