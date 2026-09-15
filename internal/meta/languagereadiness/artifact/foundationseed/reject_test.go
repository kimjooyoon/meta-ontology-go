package foundationseed

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
)

func TestUnknownOrNoncontiguousEvidenceFailsClosed(t *testing.T) {
	canonical := exactResolution(t)
	noncontiguous := canonical
	noncontiguous.Attempts = append([]predecessorresolution.AttemptReceipt(nil),
		canonical.Attempts...)
	noncontiguous.Attempts[0].ParentSHA = testSHA(999)
	for index, input := range []predecessorresolution.Report{noncontiguous, canonical} {
		expectedHead := canonical.CurrentHeadSHA
		if index == 1 {
			expectedHead = testSHA(998)
		}
		report := Evaluate(input, expectedHead)
		if err := Validate(report); err != nil {
			t.Fatal(err)
		}
		if report.Decision != DecisionFailClosed ||
			report.Resolution != ResolutionLower ||
			report.Source.ExactExhaustion {
			t.Fatalf("unknown evidence accepted: %#v", report)
		}
	}
}
