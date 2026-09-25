package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

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
