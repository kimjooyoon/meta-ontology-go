package predecessorbinding

import "testing"

const testHead = "1111111111111111111111111111111111111111"

func observations(state State) []Observation {
	result := make([]Observation, 0, Total)
	for _, coordinate := range Registry() {
		result = append(result, Observation{ID: coordinate.ID, GoField: coordinate.GoField,
			SourcePath: SourcePath, Provider: Provider, State: state})
	}
	return result
}

func TestEvaluateCountsStaticBaselineWithoutClaimingImprovement(t *testing.T) {
	report := Evaluate(testHead, observations(StateStaticLiteral), 0)
	if report.Decision != DecisionPass || report.Summary.StaticLiteral != 8 ||
		report.Summary.DynamicInput != 0 || report.Summary.Unknown != 0 ||
		report.Summary.DynamicBPS != 0 {
		t.Fatalf("unexpected static baseline: %+v", report)
	}
	if report.Indicators[0].Satisfied || report.Indicators[2].Satisfied {
		t.Fatal("static baseline must not satisfy dynamic outcome or static guardrail")
	}
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateCountsDynamicTargetOnSameRegistry(t *testing.T) {
	report := Evaluate(testHead, observations(StateDynamicInput), 0)
	if report.Decision != DecisionPass || report.Summary.StaticLiteral != 0 ||
		report.Summary.DynamicInput != 8 || report.Summary.DynamicBPS != 10_000 {
		t.Fatalf("unexpected dynamic target: %+v", report)
	}
}

func TestEvaluateFailsClosedOnMissingEvidence(t *testing.T) {
	report := Evaluate(testHead, observations(StateStaticLiteral)[:7], 0)
	if report.Decision != DecisionFailClosed || report.Reason != ReasonEvidenceUnknown ||
		report.Summary.Unknown != 1 {
		t.Fatalf("unknown evidence did not fail closed: %+v", report)
	}
}
