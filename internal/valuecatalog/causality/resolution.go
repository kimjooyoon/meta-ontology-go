package causality

func incomingUnavailableEdges(graph GraphContract, claimID string, unavailable []bool, claimIDs []string) []GraphEdge {
	indexes := make(map[string]int, len(claimIDs))
	for index, id := range claimIDs {
		indexes[id] = index
	}
	var edges []GraphEdge
	for _, edge := range graph.Edges {
		if edge.ToClaimID == claimID && unavailable[indexes[edge.FromClaimID]] {
			edges = append(edges, edge)
		}
	}
	return edges
}

func primaryCausePath(index int, unavailable []bool) []int {
	for _, edge := range edgeContract {
		if edge.to == index && unavailable[edge.from] {
			path := primaryCausePath(edge.from, unavailable)
			return append(path, index)
		}
	}
	return []int{index}
}

func missingEvidenceID(claimID string) string {
	return "evidence:" + claimID
}

func nonEmptyCoordinate(coordinate Coordinate) Coordinate {
	if coordinate.Stage == "" {
		coordinate.Stage = "RESOLVE"
	}
	if coordinate.Step == "" {
		coordinate.Step = "resolve-operation-spec"
	}
	if coordinate.Reason == "" {
		coordinate.Reason = "VALUE_PROGRAM_UNKNOWN"
	}
	return coordinate
}
