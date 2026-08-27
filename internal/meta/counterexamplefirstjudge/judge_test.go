package counterexamplefirstjudge

import (
	"testing"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
)

func TestIndependentJudgeRejectsReceiptMutation(t *testing.T) {
	contract := cf.CanonicalContract()
	scenario := cf.Scenario{ID: "resolved-minimal-counterexample",
		Candidate: cf.Candidate{ID: "candidate", SuccessExample: "success://example"},
		Counterexample: &cf.Counterexample{ID: "ce", Stage: "COUNTEREXAMPLE", Step: "shrink", Reason: "MISSING_RESOLUTION_PROBE", Failing: true, Size: 1, Minimal: true,
			ShrinkTrace: []cf.ShrinkStep{{FromSize: 2, ToSize: 1, PreservesFailure: true}}},
	}
	scenario.Resolution = &cf.ResolutionEvidence{ID: "resolution", CounterexampleID: "ce", Stage: "RESOLUTION", Step: "prove-repair", Reason: "RESOLUTION_EVIDENCE_ACCEPTED",
		ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "resolve-minimal-counterexample", Producer: "counterexample-resolution-witness", Consumer: cf.ProducerID, Accepted: true}
	spec := contract.Cases[0]
	receipt := independentlyExpectedReceipt(contract, "head", contract.SourcePath, []byte("source"), spec, scenario)
	receipt.Decision = "PASS"
	report := Evaluate(cf.JudgeInput{Contract: contract, HeadSHA: "head", SourcePath: contract.SourcePath,
		Source: []byte(`package counterexamplefirst
namespace counterexamplefirst
entity CompilationClaim id "gooo://counterexample-first/entity/compilation-claim"
entity MinimalCounterexample id "gooo://counterexample-first/entity/minimal-counterexample"
entity ResolutionEvidence id "gooo://counterexample-first/entity/resolution-evidence"
entity CompilationDecision id "gooo://counterexample-first/entity/compilation-decision"
activity DiscoverMinimalCounterexample(CompilationClaim) -> MinimalCounterexample
activity BindResolutionEvidence(MinimalCounterexample) -> ResolutionEvidence
activity PromoteOnlyAfterResolution(ResolutionEvidence) -> CompilationDecision
`), Corpus: cf.ScenarioCorpus{Schema: cf.CorpusSchema, Version: 1, Scenarios: make([]cf.Scenario, cf.CaseCount)}, Receipts: []cf.DecisionReceipt{receipt}})
	if report.Decision != "FAIL_CLOSED" || report.Reason != "COUNTEREXAMPLE_RECEIPT_COUNT_MISMATCH" {
		t.Fatalf("report=%#v", report)
	}
}
