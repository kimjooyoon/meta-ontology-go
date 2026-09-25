package externalecosystemconformance

import "testing"

func TestUnavailableEvidenceFailsClosed(t *testing.T) {
	capsule, err := Reference()
	if err != nil {
		t.Fatal(err)
	}
	report := Evaluate("test-subject", capsule, Evidence{})
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown {
		t.Fatalf("got %s/%s", report.Decision, report.Resolution)
	}
	if report.EnforcementEffect != EffectBlock || report.RepositoryWrites != 0 {
		t.Fatalf("effect=%s writes=%d", report.EnforcementEffect, report.RepositoryWrites)
	}
}

func TestUnknownRelationLowersResolution(t *testing.T) {
	capsule, err := Reference()
	if err != nil {
		t.Fatal(err)
	}
	capsule.Capabilities[0].Relation = "UNRECOGNIZED"
	resolution, reason := validateCapsule(capsule)
	if resolution != ResolutionUnknown || reason != ReasonRelationUnknown {
		t.Fatalf("got %s/%s", resolution, reason)
	}
}

func TestFixedCaseDenominator(t *testing.T) {
	if len(caseIDs) != 10 {
		t.Fatalf("case denominator=%d", len(caseIDs))
	}
	if len(capabilityRules) != 8 {
		t.Fatalf("reference denominator=%d", len(capabilityRules))
	}
}
