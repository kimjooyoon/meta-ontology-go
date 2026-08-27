package decider

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/producer"
)

func TestDecideSelectsEarliestChangedAxis(t *testing.T) {
	receipt, err := producer.BuildReceipt([]byte(`package p
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
	`), "0123456789abcdef0123456789abcdef01234567", baseContext(), model.DefaultIndependenceEvidence())
	if err != nil {
		t.Fatal(err)
	}
	context := baseContext()
	context.Tuple = receipt.Tuple
	context.Tuple.Runner += "|changed"
	context.Tuple.Subject += "|changed"
	verdict := Decide(mustJSON(receipt), mustJSON(context))
	if verdict.State != model.StateStale || verdict.Coordinate.Stage != model.StageSubject ||
		verdict.Reason != "SUBJECT_CHANGED" {
		t.Fatalf("verdict=%+v", verdict)
	}
}

func TestDecideDoesNotPreserveUnknownOrExpiredClaims(t *testing.T) {
	receipt, err := producer.BuildReceipt([]byte(`package p
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
	`), "0123456789abcdef0123456789abcdef01234567", baseContext(), model.DefaultIndependenceEvidence())
	if err != nil {
		t.Fatal(err)
	}
	unknown := baseContext()
	unknown.Tuple = receipt.Tuple
	unknown.Tuple.Verifier = ""
	unknownVerdict := Decide(mustJSON(receipt), mustJSON(unknown))
	if unknownVerdict.State != model.StateUnknown || unknownVerdict.Transition.To != "CLAIM_UNKNOWN" {
		t.Fatalf("unknown=%+v", unknownVerdict)
	}
	expired := baseContext()
	expired.Tuple = receipt.Tuple
	expired.CurrentEpoch = receipt.Boundary.ValidThroughEpoch + 1
	expiredVerdict := Decide(mustJSON(receipt), mustJSON(expired))
	if expiredVerdict.State != model.StateStale || expiredVerdict.Coordinate.Stage != model.StageVerifier ||
		expiredVerdict.Reason != "TEMPORAL_BOUNDARY_EXPIRED" || expiredVerdict.Transition.To != "CLAIM_STALE" {
		t.Fatalf("expired=%+v", expiredVerdict)
	}
}

func baseContext() model.Context {
	return model.Context{Schema: model.ContextSchema, Tuple: model.EvidenceTuple{
		Recipe: "recipe:gooo-evidence-freshness/v1", Environment: "environment:go1.27/linux-amd64/hermetic",
		Runner: "runner:github-actions/ubuntu-24.04", Verifier: "verifier:evidence-freshness-decider/v1",
	}, CurrentEpoch: 20260827, EnvironmentBoundary: "environment:go1.27/linux-amd64/hermetic", Consumer: model.ConsumerID}
}

func mustJSON(value any) []byte {
	raw, err := model.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
