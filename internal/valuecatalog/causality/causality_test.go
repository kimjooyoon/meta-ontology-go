package causality

import "testing"

func TestEvaluateSuccess(t *testing.T) {
	receipt, err := Evaluate(fixtureReport(t, ModeSuccess), ModeSuccess)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Metrics.DischargedClaimTotal != 9 || receipt.Metrics.DirectMissingClaimTotal != 0 || receipt.Metrics.DependencyBlockedClaimTotal != 0 {
		t.Fatalf("unexpected success metrics: %+v", receipt.Metrics)
	}
	if receipt.Decision.Value != "PASS" || receipt.Decision.SemanticPromotionAuthorized {
		t.Fatalf("unexpected success decision: %+v", receipt.Decision)
	}
}

func TestEvaluateUnknownSeparatesDirectAndBlocked(t *testing.T) {
	receipt, err := Evaluate(fixtureReport(t, ModeUnknown), ModeUnknown)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	metrics := receipt.Metrics
	if metrics.DirectMissingClaimTotal != 1 || metrics.DependencyBlockedClaimTotal != 8 || metrics.ObservedBlockingEdgeTotal != 11 || metrics.MaximumCausePathDepth != 4 {
		t.Fatalf("unexpected unknown metrics: %+v", metrics)
	}
	if receipt.Resolutions[0].Kind != "DIRECT_MISSING" {
		t.Fatalf("root classification: %+v", receipt.Resolutions[0])
	}
	for _, resolution := range receipt.Resolutions[1:] {
		if resolution.Kind != "DEPENDENCY_BLOCKED" || resolution.Coordinate.Reason != "UPSTREAM_CLAIM_OPEN" {
			t.Fatalf("blocked classification: %+v", resolution)
		}
	}
}
