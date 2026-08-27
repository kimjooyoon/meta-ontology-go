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
	`), Independence: model.DefaultIndependenceEvidence()})
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.CasesSatisfied != model.CaseTotal || report.Summary.FreshCases != 1 ||
		report.Summary.StaleCases != 7 || report.Summary.UnknownCases != 2 || report.Summary.AxisChangesObserved != model.AxisTotal ||
		report.Summary.ForbiddenDependencyCount != 0 || report.Summary.IndependenceContract != (model.FixedMetric{Numerator: 1, Denominator: 1}) ||
		report.Independence != model.DefaultIndependenceEvidence() || report.Receipt.Independence != report.Independence {
		t.Fatalf("summary=%+v", report.Summary)
	}
}

func TestEvaluateFailsClosedWhenForbiddenDependenciesAreObserved(t *testing.T) {
	input := Input{Contract: CanonicalContract(), HeadSHA: strings.Repeat("a", 40), Source: []byte("invalid"),
		Independence: model.IndependenceEvidence{Schema: model.IndependenceSchema, ForbiddenDependencyCount: 1,
			IndependenceContract: model.FixedMetric{Numerator: model.IndependenceContractTotal, Denominator: model.IndependenceContractTotal}}}
	report := Evaluate(input)
	if report.Decision != model.DecisionFailClosed || report.Reason != "EVIDENCE_FRESHNESS_CONTRACT_MISMATCH" {
		t.Fatalf("report=%+v", report)
	}
}
