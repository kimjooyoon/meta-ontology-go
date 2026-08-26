package sourceauthoritypromotion

import "testing"

func TestEvaluateEligibleWithoutApplyingPromotion(t *testing.T) {
	report := Evaluate(validInput(t))
	if report.Decision != DecisionEligible || report.Resolution != ResolutionExact || report.Enforcement != EnforcementNoEffect {
		t.Fatalf("unexpected eligibility: %#v", report)
	}
	if report.Summary.BeforeOperating != 6 || report.Summary.AfterOperating != 7 || report.Summary.AfterCoverageBPS != 5833 {
		t.Fatalf("unexpected transition summary: %#v", report.Summary)
	}
	if report.RepositoryWrites != 0 || report.PromotionApplied != 0 || len(report.Indicators) != 6 {
		t.Fatalf("eligibility had effects: %#v", report)
	}
}
