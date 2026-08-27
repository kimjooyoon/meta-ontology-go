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
	if receipt.Subject.BindingStatus != "PARTIAL_UNKNOWN" || !reflect.DeepEqual(receipt.Subject.MissingBindingEvidence, bindingEvidence(receipt.Subject.SourceDigest, receipt.Subject.SemanticIRDigest)) {
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
