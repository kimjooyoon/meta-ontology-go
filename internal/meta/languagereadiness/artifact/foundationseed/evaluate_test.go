package foundationseed

import "testing"

func TestExactExhaustionAuthorizesOnlyFoundationSeed(t *testing.T) {
	input := exactResolution(t)
	report := Evaluate(input, input.CurrentHeadSHA)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionAuthorized || report.Reason != ReasonExact ||
		report.Summary.Satisfied != 12 || report.Summary.Total != 12 ||
		report.Views[0].Satisfied != 3 || report.Views[0].Total != 3 ||
		report.Views[1].Satisfied != 11 || report.Views[1].Total != 11 ||
		report.Views[2].Satisfied != 12 || report.Views[2].Total != 12 {
		t.Fatalf("foundation report = %#v", report)
	}
	if !authorityDenied(report.Authority) ||
		report.Source.RepositoryWrites != 0 ||
		report.Source.ReadinessDeltaClaims != 0 {
		t.Fatalf("foundation authority = %#v", report)
	}
}
