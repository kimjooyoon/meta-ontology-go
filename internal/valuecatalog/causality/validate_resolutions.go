package causality

import "fmt"

func validateResolutions(receipt Receipt) (string, error) {
	accepted := 0
	direct := 0
	blocked := 0
	edges := make(map[string]GraphEdge, len(receipt.Graph.Edges))
	for _, edge := range receipt.Graph.Edges {
		edges[edge.EdgeID] = edge
	}
	var root Resolution
	for _, resolution := range receipt.Resolutions {
		switch resolution.Kind {
		case "EVIDENCE_ACCEPTED":
			accepted++
			if resolution.State != "DISCHARGED" || len(resolution.CausePath) != 0 {
				return "", fmt.Errorf("accepted claim %q has causal residue", resolution.ClaimID)
			}
		case "DIRECT_MISSING":
			direct++
			root = resolution
			if resolution.State != "OPEN" || len(resolution.CausePath) != 1 || resolution.CausePath[0] != resolution.ClaimID {
				return "", fmt.Errorf("direct claim %q path mismatch", resolution.ClaimID)
			}
			if resolution.CauseTransitionDigest == "" || resolution.CauseCoordinate == nil || len(resolution.MissingEvidenceIDs) != 1 {
				return "", fmt.Errorf("direct claim %q cause binding missing", resolution.ClaimID)
			}
		case "DEPENDENCY_BLOCKED":
			blocked++
			if resolution.State != "OPEN" || resolution.Coordinate.Reason != "UPSTREAM_CLAIM_OPEN" {
				return "", fmt.Errorf("blocked claim %q coordinate mismatch", resolution.ClaimID)
			}
			if len(resolution.BlockedByClaimIDs) == 0 || len(resolution.BlockedByClaimIDs) != len(resolution.BlockedByEdgeIDs) {
				return "", fmt.Errorf("blocked claim %q frontier mismatch", resolution.ClaimID)
			}
			for index, edgeID := range resolution.BlockedByEdgeIDs {
				edge, ok := edges[edgeID]
				if !ok || edge.FromClaimID != resolution.BlockedByClaimIDs[index] || edge.ToClaimID != resolution.ClaimID {
					return "", fmt.Errorf("blocked claim %q edge mismatch", resolution.ClaimID)
				}
			}
			if len(resolution.CausePath) < 2 || resolution.CausePath[len(resolution.CausePath)-1] != resolution.ClaimID {
				return "", fmt.Errorf("blocked claim %q cause path mismatch", resolution.ClaimID)
			}
			for index := 1; index < len(resolution.CausePath); index++ {
				if !hasGraphEdge(receipt.Graph, resolution.CausePath[index-1], resolution.CausePath[index]) {
					return "", fmt.Errorf("blocked claim %q path edge missing", resolution.ClaimID)
				}
			}
		default:
			return "", fmt.Errorf("unsupported resolution kind %q", resolution.Kind)
		}
	}
	switch {
	case accepted == ClaimTotal && direct == 0 && blocked == 0:
		return ModeSuccess, nil
	case accepted == 0 && direct == 1 && blocked == ClaimTotal-1:
		for _, resolution := range receipt.Resolutions {
			if resolution.Kind == "DEPENDENCY_BLOCKED" {
				if resolution.CausePath[0] != root.ClaimID || resolution.CauseTransitionDigest != root.CauseTransitionDigest || resolution.CauseCoordinate == nil || *resolution.CauseCoordinate != *root.CauseCoordinate {
					return "", fmt.Errorf("blocked claim %q root cause mismatch", resolution.ClaimID)
				}
			}
		}
		if receipt.Metrics.ObservedBlockingEdgeTotal != EdgeTotal || receipt.Metrics.MaximumCausePathDepth != 4 {
			return "", fmt.Errorf("unknown causal shape mismatch")
		}
		return ModeUnknown, nil
	default:
		return "", fmt.Errorf("resolution cardinality outside v1 contract: accepted=%d direct=%d blocked=%d", accepted, direct, blocked)
	}
}
