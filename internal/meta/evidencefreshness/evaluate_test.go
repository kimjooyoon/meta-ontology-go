package evidencefreshness

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

func TestEvaluatePinsFreshnessDenominatorsAndTransitions(t *testing.T) {
	report := Evaluate(Input{Contract: CanonicalContract(), HeadSHA: strings.Repeat("a", 40),
		Source: []byte(`package p
namespace p
entity A id "a"
entity B id "b"
entity C id "c"
entity D id "d"
entity E id "e"
entity F id "f"
activity One(A) -> B
activity Two(B) -> C
activity Three(C) -> D
`), Independence: model.IndependenceEvidence{Schema: model.IndependenceSchema}})
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.CasesSatisfied != model.CaseTotal || report.Summary.FreshCases != 1 ||
		report.Summary.StaleCases != 7 || report.Summary.UnknownCases != 2 || report.Summary.AxisChangesObserved != model.AxisTotal {
		t.Fatalf("summary=%+v", report.Summary)
	}
}

func TestEvaluateFailsClosedWhenDeciderDependenciesAreDeclared(t *testing.T) {
	input := Input{Contract: CanonicalContract(), HeadSHA: strings.Repeat("a", 40), Source: []byte("invalid"),
		Independence: model.IndependenceEvidence{Schema: model.IndependenceSchema, DeciderDependencies: 1}}
	report := Evaluate(input)
	if report.Decision != model.DecisionFailClosed || report.Reason != "EVIDENCE_FRESHNESS_CONTRACT_MISMATCH" {
		t.Fatalf("report=%+v", report)
	}
}
