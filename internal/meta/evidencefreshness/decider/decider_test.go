package decider

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/producer"
)

func TestDecideReconstructsSourceAndSelectsEarliestChangedAxis(t *testing.T) {
	source := fixtureSource()
	receipt, err := producer.BuildReceipt(source, "0123456789abcdef0123456789abcdef01234567", baseContext(), model.DefaultIndependenceEvidence(), model.DefaultWriteSetObservation())
	if err != nil {
		t.Fatal(err)
	}
	context := baseContext()
	context.PolicyDigest = receipt.PolicyDigest
	context.Tuple = receipt.Tuple
	context.Tuple.Runner += "|changed"
	context.Tuple.Subject += "|changed"
	verdict := Decide(source, mustJSON(receipt), mustJSON(context))
	if verdict.State != model.StateStale || verdict.Coordinate.Stage != model.StageSubject ||
		verdict.Reason != "SUBJECT_CHANGED" || verdict.RawFreshness != model.StateFresh || verdict.SemanticFreshness != model.StateFresh {
		t.Fatalf("verdict=%+v", verdict)
	}
}

func TestDecidePreservesCommentOnlySemanticClaim(t *testing.T) {
	source := fixtureSource()
	receipt, err := producer.BuildReceipt(source, "0123456789abcdef0123456789abcdef01234567", baseContext(), model.DefaultIndependenceEvidence(), model.DefaultWriteSetObservation())
	if err != nil {
		t.Fatal(err)
	}
	commentSource := append([]byte("// presentation-only\n"), source...)
	context := baseContext()
	context.PolicyDigest = receipt.PolicyDigest
	context.Tuple = receipt.Tuple
	context.Tuple.Material.RawDigest = model.DigestBytes(commentSource)
	verdict := Decide(commentSource, mustJSON(receipt), mustJSON(context))
	if verdict.State != model.StateFresh || verdict.Decision != model.DecisionPass || verdict.RawFreshness != model.StateStale ||
		verdict.SemanticFreshness != model.StateFresh || verdict.Reason != "RAW_MATERIAL_CHANGED_SEMANTIC_PRESERVED" {
		t.Fatalf("verdict=%+v", verdict)
	}
}

func TestDecideLowersUnavailableSourceAndExpiredClaim(t *testing.T) {
	source := fixtureSource()
	receipt, err := producer.BuildReceipt(source, "0123456789abcdef0123456789abcdef01234567", baseContext(), model.DefaultIndependenceEvidence(), model.DefaultWriteSetObservation())
	if err != nil {
		t.Fatal(err)
	}
	context := baseContext()
	context.PolicyDigest = receipt.PolicyDigest
	context.Tuple = receipt.Tuple
	unknownVerdict := Decide(nil, mustJSON(receipt), mustJSON(context))
	if unknownVerdict.State != model.StateUnknown || unknownVerdict.Resolution != model.ResolutionLower ||
		unknownVerdict.Coordinate.Stage != model.StageSubject || unknownVerdict.Coordinate.Step != "reconstruct-source" || unknownVerdict.Reason != "SOURCE_UNAVAILABLE" {
		t.Fatalf("unknown=%+v", unknownVerdict)
	}
	expired := context
	expired.CurrentEpoch = receipt.Boundary.ValidThroughEpoch + 1
	expiredVerdict := Decide(source, mustJSON(receipt), mustJSON(expired))
	if expiredVerdict.State != model.StateStale || expiredVerdict.Coordinate.Stage != model.StageVerifier ||
		expiredVerdict.Reason != "TEMPORAL_BOUNDARY_EXPIRED" || expiredVerdict.Transition.To != "CLAIM_STALE" {
		t.Fatalf("expired=%+v", expiredVerdict)
	}
}

func baseContext() model.Context {
	return model.Context{Schema: model.ContextSchema, Tuple: model.EvidenceTuple{
		Recipe: "recipe:gooo-evidence-freshness/v2", Environment: "environment:go1.27/linux-amd64/hermetic",
		Runner: "runner:github-actions/ubuntu-24.04", Verifier: "verifier:evidence-freshness-decider/v2",
	}, CurrentEpoch: 20260827, EnvironmentBoundary: "environment:go1.27/linux-amd64/hermetic", Consumer: model.ConsumerID}
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
entity A id "a"
entity B id "b"
entity C id "c"
entity D id "d"
entity E id "e"
entity F id "f"
activity One(A) -> B
activity Two(B) -> C
activity Three(C) -> D
`)
}

func mustJSON(value any) []byte {
	raw, err := model.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
