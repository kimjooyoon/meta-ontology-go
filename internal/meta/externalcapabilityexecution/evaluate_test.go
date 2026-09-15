package externalcapabilityexecution

import "testing"

func TestUnknownParentDecisionFailsClosed(t *testing.T) {
	report := RunCase("subject", "parent-decision-unknown")
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown {
		t.Fatalf("got %s/%s", report.Decision, report.Resolution)
	}
}

func TestExactCapabilityDoesNotPromoteParent(t *testing.T) {
	report := RunCase("subject", "exact")
	if report.Decision != DecisionExecutable || report.PromotionCount != 0 {
		t.Fatalf("decision=%s promotions=%d", report.Decision, report.PromotionCount)
	}
	if report.Parent.Decision != DecisionFailClosed || report.Parent.Completed != 6 {
		t.Fatalf("parent=%s completed=%d", report.Parent.Decision, report.Parent.Completed)
	}
}

func TestFixedDenominators(t *testing.T) {
	if MetricDenominator != 10 || len(caseIDs) != SuiteDenominator {
		t.Fatalf("metrics=%d cases=%d", MetricDenominator, len(caseIDs))
	}
}
