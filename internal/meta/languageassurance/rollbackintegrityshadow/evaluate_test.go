package rollbackintegrityshadow

import "testing"

func TestExactShadowBindsExistingRollbackMetaCode(t *testing.T) {
	report := Evaluate(EmbeddedAssurance())
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionShadowPass || report.Summary.CasesPassed != 7 ||
		report.Summary.CoordinatesTotal != 70 || report.Summary.ProjectedOperating != 10 ||
		report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		t.Fatalf("report = %#v", report)
	}
	for _, result := range report.Cases {
		if result.Name == "unknown-guard-decision" &&
			(result.ActualDecision != DecisionFailClosed || result.ActualResolution != ResolutionLower) {
			t.Fatalf("unknown case = %#v", result)
		}
	}
}

func TestUnavailableAssuranceLowersResolution(t *testing.T) {
	report := Evaluate(nil)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionLower ||
		report.Reason != ReasonUnavailable || report.Summary.ProjectedOperating != 9 {
		t.Fatalf("report = %#v", report)
	}
}

func TestAssuranceDigestMismatchPreservesInvariantOnly(t *testing.T) {
	raw := append(EmbeddedAssurance(), '\n')
	report := Evaluate(raw)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionInvariant ||
		report.Reason != ReasonDigest || report.PromotionApplied != 0 {
		t.Fatalf("report = %#v", report)
	}
}
