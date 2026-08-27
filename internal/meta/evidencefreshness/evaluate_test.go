package evidencefreshness

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

func TestEvaluatePinsFreshnessAndLedgerDenominators(t *testing.T) {
	report := Evaluate(Input{Contract: CanonicalContract(), HeadSHA: strings.Repeat("a", 40), Source: fixtureSource(),
		Independence: model.DefaultIndependenceEvidence(), WriteSet: model.DefaultWriteSetObservation()})
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.CasesObserved != model.CaseTotal || report.Summary.CurrentEvidenceCases != model.CurrentEvidenceTotal ||
		report.Summary.SyntheticCounterexamples != model.SyntheticCounterexampleTotal || report.Summary.RawStaleCases != 2 ||
		report.Summary.SemanticStaleCases != 1 || report.Summary.ClaimDischarged != 2 || report.Summary.ClaimOpenPreserved != 8 ||
		report.Summary.SourceReconstructedCases != 9 || report.Summary.SourceUnavailableCases != 1 {
		t.Fatalf("summary=%+v", report.Summary)
	}
	comment, semantic, unavailable := caseByID(report.Cases, "synthetic-comment-only"), caseByID(report.Cases, "synthetic-semantic-change"), caseByID(report.Cases, "synthetic-source-unavailable")
	if comment.RawFreshness != model.StateStale || comment.SemanticFreshness != model.StateFresh || comment.ObservedDecision != model.DecisionPass ||
		semantic.SemanticFreshness != model.StateStale || semantic.ObservedDecision != model.DecisionFailClosed ||
		unavailable.ObservedState != model.StateUnknown || unavailable.ObservedResolution != model.ResolutionLower {
		t.Fatalf("interventions comment=%+v semantic=%+v unavailable=%+v", comment, semantic, unavailable)
	}
}

func TestEvaluateFailsClosedWhenForbiddenDependencyIsObserved(t *testing.T) {
	input := Input{Contract: CanonicalContract(), HeadSHA: strings.Repeat("a", 40), Source: fixtureSource(),
		Independence: model.IndependenceEvidence{Schema: model.IndependenceSchema, ForbiddenDependencyCount: 1,
			IndependenceContract: model.FixedMetric{Numerator: model.IndependenceContractTotal, Denominator: model.IndependenceContractTotal}},
		WriteSet: model.DefaultWriteSetObservation()}
	report := Evaluate(input)
	if report.Decision != model.DecisionFailClosed || report.Reason != "EVIDENCE_FRESHNESS_CONTRACT_MISMATCH" ||
		report.Summary.ForbiddenDependencyCount != 1 || report.Summary.IndependenceContract.Numerator != 1 {
		t.Fatalf("report=%+v", report)
	}
}

func fixtureSource() []byte {
	return []byte(`package p
namespace p
freshness axes subject material recipe environment runner verifier
freshness comparison_policy earliest_changed
freshness prior_claim_state OPEN
freshness boundary_policy logical_epoch_environment
freshness raw_material_policy raw_material_digest
freshness semantic_policy comments_ignored
freshness claim_ledger_policy open_discharge_or_preserve
freshness effect_policy read_only_ci_before_after
entity A id "gooo://evidence-freshness/test/a"
entity B id "gooo://evidence-freshness/test/b"
entity C id "gooo://evidence-freshness/test/c"
entity D id "gooo://evidence-freshness/test/d"
entity E id "gooo://evidence-freshness/test/e"
entity F id "gooo://evidence-freshness/test/f"
activity One(A) -> B
activity Two(B) -> C
activity Three(C) -> D
`)
}
