package counterexamplefirstcompiler

import (
	"testing"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
)

const sourceFixture = `package counterexamplefirst
namespace counterexamplefirst

entity CompilationClaim id "gooo://counterexample-first/entity/compilation-claim"
entity MinimalCounterexample id "gooo://counterexample-first/entity/minimal-counterexample"
entity ResolutionEvidence id "gooo://counterexample-first/entity/resolution-evidence"
entity CompilationDecision id "gooo://counterexample-first/entity/compilation-decision"

activity DiscoverMinimalCounterexample(CompilationClaim) -> MinimalCounterexample
activity BindResolutionEvidence(MinimalCounterexample) -> ResolutionEvidence
activity PromoteOnlyAfterResolution(ResolutionEvidence) -> CompilationDecision
`

func TestCompileBlocksSuccessExampleWithoutCounterexample(t *testing.T) {
	contract := cf.CanonicalContract()
	scenarios := validCorpus()
	scenarios.Scenarios[1].Counterexample = nil
	scenarios.Scenarios[1].Resolution = nil
	receipts, err := Compile(contract, "head", contract.SourcePath, []byte(sourceFixture), scenarios)
	if err != nil {
		t.Fatal(err)
	}
	for _, receipt := range receipts {
		if receipt.ScenarioID == "success-example-only" &&
			(receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "COUNTEREXAMPLE_REQUIRED") {
			t.Fatalf("receipt=%#v", receipt)
		}
	}
}

func TestCompilePromotesOnlyAfterResolvedMinimalCounterexample(t *testing.T) {
	receipts, err := Compile(cf.CanonicalContract(), "head", cf.CanonicalContract().SourcePath,
		[]byte(sourceFixture), validCorpus())
	if err != nil {
		t.Fatal(err)
	}
	if receipts[0].Decision != "PASS" || receipts[0].Coordinate.Stage != "COMPILE_DECISION" ||
		receipts[0].DecisionInput.CounterexampleID == "" || receipts[0].DecisionInput.ResolutionID == "" {
		t.Fatalf("receipt=%#v", receipts[0])
	}
}

func TestCompilePreservesUnknownCoordinate(t *testing.T) {
	scenarios := validCorpus()
	scenarios.Scenarios[4].Counterexample.Stage = "UNKNOWN"
	scenarios.Scenarios[4].Counterexample.Step = "UNKNOWN"
	scenarios.Scenarios[4].Counterexample.Reason = "UNKNOWN"
	receipts, err := Compile(cf.CanonicalContract(), "head", cf.CanonicalContract().SourcePath,
		[]byte(sourceFixture), scenarios)
	if err != nil {
		t.Fatal(err)
	}
	if receipts[4].Decision != "UNKNOWN" || receipts[4].Coordinate != (cf.Coordinate{Stage: "UNKNOWN", Step: "UNKNOWN", Reason: "UNKNOWN"}) {
		t.Fatalf("receipt=%#v", receipts[4])
	}
}

func validCorpus() cf.ScenarioCorpus {
	contract := cf.CanonicalContract()
	result := cf.ScenarioCorpus{Schema: cf.CorpusSchema, Version: 1, Scenarios: make([]cf.Scenario, len(contract.Cases))}
	for index, spec := range contract.Cases {
		counterexample := cf.Counterexample{ID: "ce-" + spec.ID, Stage: "COUNTEREXAMPLE", Step: "shrink", Reason: "MISSING_RESOLUTION_PROBE", Failing: true, Size: 1, Minimal: true,
			ShrinkTrace: []cf.ShrinkStep{{FromSize: 2, ToSize: 1, PreservesFailure: true}}}
		result.Scenarios[index] = cf.Scenario{ID: spec.ID,
			Candidate:      cf.Candidate{ID: "candidate-" + spec.ID, Claim: "claim", SuccessExample: "success://example"},
			Counterexample: &counterexample}
		if spec.ID == "resolved-minimal-counterexample" {
			result.Scenarios[index].Resolution = &cf.ResolutionEvidence{ID: "resolution-" + spec.ID,
				CounterexampleID: counterexample.ID, Stage: "RESOLUTION", Step: "prove-repair",
				Reason: "RESOLUTION_EVIDENCE_ACCEPTED", ProofChoice: "COUNTEREXAMPLE_RESOLUTION",
				MetaOperation: "resolve-minimal-counterexample", Producer: "counterexample-resolution-witness",
				Consumer: cf.ProducerID, Accepted: true}
		}
	}
	return result
}
