package toolchainconformance

import "testing"

func TestEvaluateFailsClosedOnMissingSurface(t *testing.T) {
	input := fixtureInput(t)
	delete(input.Artifacts, "toolchain-cli")
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed ||
		report.Resolution != ResolutionLower ||
		report.Summary.MissingSurfaces != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestValidateRejectsUnknownTopDecision(t *testing.T) {
	report := Evaluate(fixtureInput(t))
	report.Decision = "UNKNOWN"
	report = seal(report)
	if err := Validate(report, fixtureHead); err == nil {
		t.Fatal("unknown top decision was accepted")
	}
}

func TestEvaluateFailsClosedOnRegistryDrift(t *testing.T) {
	input := fixtureInput(t)
	input.RegistryRaw = []byte(`{"schema":"unknown"}`)
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed || report.Summary.RegistryDrift != 1 {
		t.Fatalf("report = %#v", report)
	}
}
