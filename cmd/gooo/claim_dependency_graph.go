package main

func claimDependencyCycleResidue(nodes []claimDependencyNode, edges []claimDependencyEdge) []string {
	indegree := make(map[string]int, len(nodes))
	outgoing := make(map[string][]string, len(nodes))
	for _, node := range nodes {
		indegree[node.Activity] = 0
	}
	for _, edge := range edges {
		indegree[edge.ToActivity]++
		outgoing[edge.FromActivity] = append(outgoing[edge.FromActivity], edge.ToActivity)
	}
	queue := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if indegree[node.Activity] == 0 {
			queue = append(queue, node.Activity)
		}
	}
	processed := 0
	for len(queue) > 0 {
		activity := queue[0]
		queue = queue[1:]
		processed++
		for _, dependent := range outgoing[activity] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}
	if processed == len(nodes) {
		return []string{}
	}
	residue := make([]string, 0, len(nodes)-processed)
	for _, node := range nodes {
		if indegree[node.Activity] > 0 {
			residue = append(residue, node.Activity)
		}
	}
	return residue
}
