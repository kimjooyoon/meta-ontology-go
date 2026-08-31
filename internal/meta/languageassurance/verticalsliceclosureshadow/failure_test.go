package verticalsliceclosureshadow

import "testing"

func TestBrokenSyntaxSemanticLinkBlocksProjection(t *testing.T) {
	input := exactInput(t)
	input.Artifacts["semantics"] = mutateFixture(t, input.Artifacts["semantics"],
		func(value map[string]any) {
			source := value["source"].(map[string]any)
			source["syntax_artifact_digest"] = fixtureDigest("0")
		})
	report := Evaluate(input)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionInvariant ||
		report.Summary.BlockedBoundaries == 0 || report.Summary.ProjectedOperating != 10 {
		t.Fatalf("report = %#v", report)
	}
}

func TestObserverWriteBlocksProjection(t *testing.T) {
	input := exactInput(t)
	input.Artifacts["use-cases"] = mutateFixture(t, input.Artifacts["use-cases"],
		func(value map[string]any) { value["repository_writes"] = float64(1) })
	report := Evaluate(input)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Resolution != ResolutionInvariant ||
		report.Summary.ObservedRepositoryWrites != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestUnavailableAssuranceLowersResolution(t *testing.T) {
	input := exactInput(t)
	input.Assurance = nil
	report := Evaluate(input)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Resolution != ResolutionLower || report.Reason != ReasonAssuranceMissing {
		t.Fatalf("report = %#v", report)
	}
}

func TestDenominatorDriftPreservesInvariantOnly(t *testing.T) {
	report := evaluate(exactInput(t), append(EmbeddedDenominator(), '\n'))
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Resolution != ResolutionInvariant || report.Reason != ReasonDenominator {
		t.Fatalf("report = %#v", report)
	}
}
