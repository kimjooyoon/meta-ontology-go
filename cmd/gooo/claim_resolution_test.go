package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestClaimResolutionTupleStates(t *testing.T) {
	cases := []struct {
		name, program, decision, state, reason string
	}{
		{"closed", "claim.resolve:v1;state=CLOSED;stage=NONE;step=NONE;reason=CLAIM_CLOSED;unknown_class=NONE;next_operation=NONE", claimDecisionObserved, claimStateClosed, "CLAIM_CLOSED"},
		{"unknown", "claim.resolve:v1;state=UNKNOWN;stage=SOURCE;step=OBSERVE_INPUT;reason=INPUT_UNAVAILABLE;unknown_class=DIRECT_MISSING;next_operation=PROVIDE_INPUT", claimDecisionObserved, claimStateUnknown, "INPUT_UNAVAILABLE"},
		{"refuted", "claim.resolve:v1;state=REFUTED;stage=SOURCE;step=VERIFY_INPUT;reason=INPUT_MISMATCH;unknown_class=NONE;next_operation=RESTORE_INPUT", claimDecisionObserved, claimStateRefuted, "INPUT_MISMATCH"},
		{"incomplete-unknown", "claim.resolve:v1;state=UNKNOWN;stage=SOURCE;step=OBSERVE_INPUT;reason=INPUT_UNAVAILABLE;unknown_class=NONE;next_operation=PROVIDE_INPUT", claimDecisionFailed, claimStateRefuted, "UNKNOWN_TUPLE_INCOMPLETE"},
		{"unknown-state", "claim.resolve:v1;state=FIXED_POINT;stage=NONE;step=NONE;reason=UNRECOGNIZED_STATE;unknown_class=NONE;next_operation=NONE", claimDecisionFailed, claimStateRefuted, "CLAIM_STATE_UNKNOWN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := fmt.Sprintf("package claims\nnamespace claims\nentity Claim id \"gooo://claim\"\nactivity Resolve(Claim) -> Claim computes %q\n", tc.program)
			file, diagnostics := syntax.ParseFile("main.gooo", source)
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics.Error())
			}
			report := resolveClaimTuple("main.gooo", []byte(source), file, "Resolve")
			if report.Decision != tc.decision || report.Claim.State != tc.state || report.Claim.Reason != tc.reason {
				t.Fatalf("resolution changed: %#v", report)
			}
		})
	}
}

func TestClaimResolutionCommandPreservesUnknownAndFailsClosed(t *testing.T) {
	valid := "package claims\nnamespace claims\nentity Claim id \"gooo://claim\"\nactivity Resolve(Claim) -> Claim computes \"claim.resolve:v1;state=UNKNOWN;stage=SOURCE;step=OBSERVE_INPUT;reason=INPUT_UNAVAILABLE;unknown_class=DIRECT_MISSING;next_operation=PROVIDE_INPUT\"\n"
	invalid := "package claims\nnamespace claims\nentity Claim id \"gooo://claim\"\nactivity Resolve(Claim) -> Claim computes \"claim.resolve:v1;state=FIXED_POINT;stage=NONE;step=NONE;reason=UNRECOGNIZED_STATE;unknown_class=NONE;next_operation=NONE\"\n"
	for _, tc := range []struct {
		name, source string
		code         int
	}{
		{"valid-unknown", valid, exitOK},
		{"invalid-state", invalid, exitFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runClaimResolution([]string{"resolve", "main.gooo", "--activity", "Resolve", "--json"}, fixtureReader{source: tc.source}, SyntaxSourceParser{}, &stdout, &stderr)
			if code != tc.code || stdout.Len() == 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestClaimResolutionExampleDenominator(t *testing.T) {
	source, err := os.ReadFile("testdata/claim-resolution/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(source, []byte(claimResolutionCandidateID)) {
		t.Fatal("candidate id is not directly mapped in Gooo source")
	}
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	valid, refuted := 0, 0
	for _, activity := range []string{"ClosedClaim", "UnknownDirectClaim", "UnknownContextClaim", "RefutedClaim", "InvalidUnknownClaim", "InvalidStateClaim"} {
		report := resolveClaimTuple("main.gooo", source, file, activity)
		if report.Decision == claimDecisionObserved {
			valid++
		} else if report.Decision == claimDecisionFailed {
			refuted++
		}
	}
	if valid != 4 || refuted != 2 {
		t.Fatalf("example denominator changed: valid=%d refuted=%d", valid, refuted)
	}
}
