package causality

func hasGraphEdge(graph GraphContract, from, to string) bool {
	for _, edge := range graph.Edges {
		if edge.FromClaimID == from && edge.ToClaimID == to {
			return true
		}
	}
	return false
}
