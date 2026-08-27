package causality

import (
	"fmt"
	"reflect"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || receipt.Scope != ReceiptScope {
		return fmt.Errorf("receipt identity mismatch")
	}
	if len(receipt.Resolutions) != ClaimTotal {
		return fmt.Errorf("resolution total: got %d want %d", len(receipt.Resolutions), ClaimTotal)
	}
	claimIDs := make([]string, ClaimTotal)
	for index, resolution := range receipt.Resolutions {
		claimIDs[index] = resolution.ClaimID
		if resolution.Axis != claimAxes[index] {
			return fmt.Errorf("resolution %d axis mismatch", index+1)
		}
	}
	expectedGraph, err := buildGraph(claimIDs)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(receipt.Graph, expectedGraph) {
		return fmt.Errorf("graph contract mismatch")
	}
	if receipt.Subject.GraphDigest != receipt.Graph.Digest || receipt.Subject.InputReportSchema != InputReportSchema {
		return fmt.Errorf("subject binding mismatch")
	}
	if receipt.Subject.InputReportDigest == "" || receipt.Subject.TransitionHeadDigest == "" {
		return fmt.Errorf("subject digest missing")
	}
	if receipt.Subject.BindingStatus != "UNKNOWN" || len(receipt.Subject.MissingBindingEvidence) != 3 {
		return fmt.Errorf("source/IR binding boundary mismatch")
	}
	expectedMetrics := deriveMetrics(receipt.Graph, receipt.Resolutions)
	if receipt.Metrics != expectedMetrics {
		return fmt.Errorf("causality metrics mismatch")
	}
	if receipt.Metrics.ClassifiedClaimTotal != ClaimTotal || receipt.Metrics.ClassificationBasisPoints != 10000 {
		return fmt.Errorf("claim classification incomplete")
	}
	expectedIndicators := buildIndicators(expectedMetrics)
	if len(receipt.Indicators) != IndicatorTotal || !reflect.DeepEqual(receipt.Indicators, expectedIndicators) {
		return fmt.Errorf("indicator contract mismatch")
	}
	mode, err := validateResolutions(receipt)
	if err != nil {
		return err
	}
	if receipt.Decision != decisionFor(mode, receipt.Metrics) {
		return fmt.Errorf("decision mismatch")
	}
	if receipt.Decision.SemanticPromotionAuthorized || receipt.Graph.SemanticCorrectnessClaimed {
		return fmt.Errorf("causal receipt cannot authorize semantic promotion")
	}
	digest, err := receiptDigest(receipt)
	if err != nil {
		return err
	}
	if receipt.ReceiptDigest != digest {
		return fmt.Errorf("receipt digest mismatch")
	}
	return nil
}

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

func hasGraphEdge(graph GraphContract, from, to string) bool {
	for _, edge := range graph.Edges {
		if edge.FromClaimID == from && edge.ToClaimID == to {
			return true
		}
	}
	return false
}
