package formatfix

import "testing"

func TestMalformedSourceLowersResolution(t *testing.T) {
	plan := Build("malformed.gooo", "package billing\nnamespace\n")
	if err := Validate(plan); err != nil {
		t.Fatal(err)
	}
	if plan.Decision != DecisionFailClosed || plan.Resolution != ResolutionLower ||
		plan.ReasonCode != "FORMAT_FIX_SOURCE_UNKNOWN" || len(plan.Edits) != 0 {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestUnknownDecisionCannotApply(t *testing.T) {
	plan := Build("billing.gooo", unformattedFixture)
	plan.Decision = "UNKNOWN"
	plan = seal(plan)
	if Validate(plan) == nil {
		t.Fatal("unknown decision accepted")
	}
	if _, err := Apply(unformattedFixture, plan); err == nil {
		t.Fatal("unknown decision applied")
	}
}
