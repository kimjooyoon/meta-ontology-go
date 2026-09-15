package verticalsliceclosureshadow

import "testing"

func TestExactEvidenceClosesSixBoundaries(t *testing.T) {
	report := Evaluate(exactInput(t))
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionShadowPass ||
		report.Summary.BoundariesSatisfied != 6 || report.Summary.LinksSatisfied != 12 ||
		report.Summary.BeforeOperating != 10 || report.Summary.ProjectedOperating != 11 ||
		report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestUnavailableBoundaryLowersResolution(t *testing.T) {
	input := exactInput(t)
	delete(input.Artifacts, "semantics")
	report := Evaluate(input)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionLower ||
		report.Summary.UnknownBoundaries == 0 || report.Summary.ProjectedOperating != 10 {
		t.Fatalf("report = %#v", report)
	}
}

func TestUnknownDecisionIsNeverAccepted(t *testing.T) {
	input := exactInput(t)
	input.Artifacts["semantics"] = mutateFixture(t, input.Artifacts["semantics"],
		func(value map[string]any) { value["decision"] = "FUTURE_DECISION" })
	report := Evaluate(input)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Resolution != ResolutionLower || report.Summary.UnknownTopDecisions != 1 {
		t.Fatalf("report = %#v", report)
	}
}
