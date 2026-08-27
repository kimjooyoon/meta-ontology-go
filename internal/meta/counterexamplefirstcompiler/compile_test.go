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

activity CanonicalEntityID(CompilationClaim) -> CompilationClaim computes "identity:v1"
activity DiscoverMinimalCounterexample(CompilationClaim) -> MinimalCounterexample
activity BindResolutionEvidence(MinimalCounterexample) -> ResolutionEvidence
activity PromoteOnlyAfterResolution(ResolutionEvidence) -> CompilationDecision
`

func TestCompileUsesRawObservationInsteadOfCorpusConclusions(t *testing.T) {
	contract := cf.CanonicalContract()
	source := sourceFixture
	corpus := cf.ScenarioCorpus{Schema: cf.CorpusSchema, Version: 1, Scenarios: []cf.Scenario{
		{ID: "resolved-minimal-counterexample", Candidate: cf.Candidate{ID: "candidate", ClaimID: "claim-resolved-repair", PredicateID: "identity-drift-detected", Claim: "the candidate identity drift can be repaired by canonicalizing the same minimal source", Source: &source}},
		{ID: "canonical-control", Candidate: cf.Candidate{ID: "control", ClaimID: "claim-canonical-control", PredicateID: "canonical-source-admissible", Claim: "a canonical candidate has no observed identity violation but is insufficient evidence for promotion", Source: &source}},
		{ID: "unresolved-counterexample", Candidate: cf.Candidate{ID: "unresolved", ClaimID: "claim-unresolved-boundary", PredicateID: "resolution-required", Claim: "an identity violation without repair evidence remains refuted", Source: &source}},
		{ID: "comment-only-control", Candidate: cf.Candidate{ID: "comment", ClaimID: "claim-comment-invariance", PredicateID: "semantic-digest-invariant", Claim: "comment-only source changes preserve semantic lowering and predicate evidence", Source: &source}},
		{ID: "unobserved-input", Candidate: cf.Candidate{ID: "missing", ClaimID: "claim-source-acquisition", PredicateID: "source-acquisition-present", Claim: "absent source acquisition retains UNKNOWN rather than inferring a decision", Source: nil}},
	}}
	receipts, err := Compile(contract, "head", contract.SourcePath, []byte(source), corpus)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipts) != cf.CaseCount {
		t.Fatalf("receipts=%d", len(receipts))
	}
	if receipts[1].Decision != "FAIL_CLOSED" || receipts[1].Reason != "COUNTEREXAMPLE_REQUIRED" {
		t.Fatalf("canonical control=%#v", receipts[1])
	}
	if receipts[4].Decision != "UNKNOWN" || receipts[4].Coordinate.Stage != "INPUT_OBSERVATION" {
		t.Fatalf("unknown input=%#v", receipts[4])
	}
}

func TestCompileRequiresResolutionRerunBeforePromotion(t *testing.T) {
	contract := cf.CanonicalContract()
	good := sourceFixture
	bad := sourceFixture
	bad = `package counterexamplefirst
namespace counterexamplefirst
entity CompilationClaim id "gooo://counterexample-first/entity/compilation-claim?drift=1"
entity MinimalCounterexample id "gooo://counterexample-first/entity/minimal-counterexample"
entity ResolutionEvidence id "gooo://counterexample-first/entity/resolution-evidence"
entity CompilationDecision id "gooo://counterexample-first/entity/compilation-decision"
activity CanonicalEntityID(CompilationClaim) -> CompilationClaim computes "identity:v1"
activity DiscoverMinimalCounterexample(CompilationClaim) -> MinimalCounterexample
activity BindResolutionEvidence(MinimalCounterexample) -> ResolutionEvidence
activity PromoteOnlyAfterResolution(ResolutionEvidence) -> CompilationDecision
`
	corpus := cf.ScenarioCorpus{Schema: cf.CorpusSchema, Version: 1, Scenarios: []cf.Scenario{
		{ID: "resolved-minimal-counterexample", Candidate: cf.Candidate{ID: "candidate", ClaimID: "claim-resolved-repair", PredicateID: "identity-drift-detected", Claim: "the candidate identity drift can be repaired by canonicalizing the same minimal source", Source: &bad}},
		{ID: "canonical-control", Candidate: cf.Candidate{ID: "control", ClaimID: "claim-canonical-control", PredicateID: "canonical-source-admissible", Claim: "a canonical candidate has no observed identity violation but is insufficient evidence for promotion", Source: &good}},
		{ID: "unresolved-counterexample", Candidate: cf.Candidate{ID: "unresolved", ClaimID: "claim-unresolved-boundary", PredicateID: "resolution-required", Claim: "an identity violation without repair evidence remains refuted", Source: &bad}},
		{ID: "comment-only-control", Candidate: cf.Candidate{ID: "comment", ClaimID: "claim-comment-invariance", PredicateID: "semantic-digest-invariant", Claim: "comment-only source changes preserve semantic lowering and predicate evidence", Source: &good}},
		{ID: "unobserved-input", Candidate: cf.Candidate{ID: "missing", ClaimID: "claim-source-acquisition", PredicateID: "source-acquisition-present", Claim: "absent source acquisition retains UNKNOWN rather than inferring a decision", Source: nil}},
	}}
	receipts, err := Compile(contract, "head", contract.SourcePath, []byte(sourceFixture), corpus)
	if err != nil {
		t.Fatal(err)
	}
	if receipts[0].Decision != "REFUTED" || receipts[0].Resolution != "LOWER_RESOLUTION" {
		t.Fatalf("unresolved promotion=%#v", receipts[0])
	}
}
