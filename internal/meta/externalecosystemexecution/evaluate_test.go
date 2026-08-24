package externalecosystemexecution

import "testing"

func TestFixedDenominator(t *testing.T) {
	if got := len(Criteria()); got != 8 { t.Fatalf("criteria=%d, want 8", got) }
	if DenominatorDigest() == "" { t.Fatal("denominator digest is empty") }
}

func TestConformanceSuite(t *testing.T) {
	suite := RunSuite()
	if suite.Passed != 10 || suite.Total != 10 || suite.Unresolved != 0 {
		t.Fatalf("suite=%d/%d unresolved=%d", suite.Passed, suite.Total, suite.Unresolved)
	}
}

func TestUnknownReferenceDecisionFailsClosed(t *testing.T) {
	observation := exactObservation()
	observation.Reference.Decision = "FIXED_POINTISH"
	report := Evaluate(&observation)
	if report.Decision != DecisionFailClosed || report.Resolution != "COARSE" {
		t.Fatalf("decision=%s resolution=%s", report.Decision, report.Resolution)
	}
}
