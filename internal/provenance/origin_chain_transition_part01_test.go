package provenance

import "testing"

func TestUpdateOriginChainRejectsTamperingAndReportsMetricChange(t *testing.T) {
	before := ObserveOriginChain(OriginChain{
		DeclarationURI:        "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol:     "activity TypedChain",
		IRNode:                "activity:TypedChain",
		GeneratedURI:          "cmd/gooo/run_source_typed_chain_part01_test.go",
		GeneratedSymbol:       "TestTypedChain",
		ReverseObservationURI: "internal/valueexecution/replay.go",
		ReverseObservation:    "execution.replay.completed",
		MetricName:            "gooo.provenance.origin-chain.v1",
		MetricValue:           "complete=1",
		EvidenceDigest:        "sha256:typed-chain-observation",
	})
	after := before
	after.Chain.EvidenceDigest = "sha256:updated-chain-observation"
	after = ObserveOriginChain(after.Chain)

	if !before.Verified() || !after.Verified() {
		t.Fatal("canonical observations must verify")
	}
	update := UpdateOriginChain(before, after)
	if update.Transition != OriginChainTransitionChanged || len(update.ChangedStages) != 1 || update.ChangedStages[0] != OriginChainStageMetric {
		t.Fatalf("metric update=%#v", update)
	}

	tampered := before
	tampered.Digest = "sha256:caller-supplied"
	if tampered.Verified() {
		t.Fatal("caller-supplied digest was accepted")
	}
	if update := UpdateOriginChain(tampered, after); update.Transition != OriginChainTransitionUnknown {
		t.Fatalf("tampered update=%#v", update)
	}
}

func TestUpdateOriginChainDoesNotComparePartialEvidence(t *testing.T) {
	partial := ObserveOriginChain(OriginChain{
		DeclarationURI:    "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol: "activity TypedChain",
		IRNode:            "activity:TypedChain",
	})
	if !partial.Verified() {
		t.Fatal("partial canonical observation must still verify its own shape")
	}
	update := UpdateOriginChain(partial, partial)
	if update.Transition != OriginChainTransitionUnknown || update.Reason != "origin comparison requires complete observations" {
		t.Fatalf("partial update=%#v", update)
	}
}
