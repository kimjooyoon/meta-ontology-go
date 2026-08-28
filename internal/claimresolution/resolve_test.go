package claimresolution

import (
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestClaimResolutionTupleStates(t *testing.T) {
	cases := []struct {
		name     string
		program  string
		decision string
		state    string
		reason   string
	}{
		{"closed", "claim.resolve:v1;state=CLOSED;stage=NONE;step=NONE;reason=CLAIM_CLOSED;unknown_class=NONE;next_operation=NONE", DecisionObserved, StateClosed, "CLAIM_CLOSED"},
		{"unknown", "claim.resolve:v1;state=UNKNOWN;stage=SOURCE;step=OBSERVE_INPUT;reason=INPUT_UNAVAILABLE;unknown_class=DIRECT_MISSING;next_operation=PROVIDE_INPUT", DecisionObserved, StateUnknown, "INPUT_UNAVAILABLE"},
		{"refuted", "claim.resolve:v1;state=REFUTED;stage=SOURCE;step=VERIFY_INPUT;reason=INPUT_MISMATCH;unknown_class=NONE;next_operation=RESTORE_INPUT", DecisionObserved, StateRefuted, "INPUT_MISMATCH"},
		{"incomplete-unknown", "claim.resolve:v1;state=UNKNOWN;stage=SOURCE;step=OBSERVE_INPUT;reason=INPUT_UNAVAILABLE;unknown_class=NONE;next_operation=PROVIDE_INPUT", DecisionFailed, StateRefuted, "UNKNOWN_TUPLE_INCOMPLETE"},
		{"unknown-state", "claim.resolve:v1;state=FIXED_POINT;stage=NONE;step=NONE;reason=UNRECOGNIZED_STATE;unknown_class=NONE;next_operation=NONE", DecisionFailed, StateRefuted, "CLAIM_STATE_UNKNOWN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := fmt.Sprintf("package claims\nnamespace claims\nentity Claim id \"gooo://claim\"\nactivity Resolve(Claim) -> Claim computes %q\n", tc.program)
			file, diagnostics := syntax.ParseFile("main.gooo", source)
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics.Error())
			}
			report := Resolve("main.gooo", []byte(source), file, "Resolve")
			if report.Decision != tc.decision || report.Claim.State != tc.state || report.Claim.Reason != tc.reason {
				t.Fatalf("resolution changed: %#v", report)
			}
		})
	}
}

func TestClaimResolutionActivityCardinalityFailsClosed(t *testing.T) {
	source := "package claims\nnamespace claims\nentity Claim id \"gooo://claim\"\n"
	file, diagnostics := syntax.ParseFile("main.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	report := Resolve("main.gooo", []byte(source), file, "Resolve")
	if report.Decision != DecisionFailed || report.Claim.Reason != "CLAIM_ACTIVITY_CARDINALITY_INVALID" {
		t.Fatalf("missing activity was accepted: %#v", report)
	}
}
