package provenance

import "testing"

func TestObserveOriginChainTracksCompleteRuntimeBoundary(t *testing.T) {
	chain := OriginChain{
		DeclarationURI:        " examples/language-runtime-binding/typed-chain.gooo ",
		DeclarationSymbol:     "activity TypedChain",
		IRNode:                "activity:TypedChain",
		GeneratedURI:          "cmd/gooo/run_source_typed_chain_part01_test.go",
		GeneratedSymbol:       "TestTypedChain",
		ReverseObservationURI: "internal/valueexecution/replay.go",
		ReverseObservation:    "execution.replay.completed",
		MetricName:            "gooo.provenance.origin-chain.v1",
		MetricValue:           "complete=1",
		EvidenceDigest:        "sha256:typed-chain-observation",
	}

	first := ObserveOriginChain(chain)
	second := ObserveOriginChain(OriginChain{
		DeclarationURI:        "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol:     " activity TypedChain ",
		IRNode:                " activity:TypedChain ",
		GeneratedURI:          "cmd/gooo/run_source_typed_chain_part01_test.go",
		GeneratedSymbol:       "TestTypedChain",
		ReverseObservationURI: "internal/valueexecution/replay.go",
		ReverseObservation:    "execution.replay.completed",
		MetricName:            "gooo.provenance.origin-chain.v1",
		MetricValue:           "complete=1",
		EvidenceDigest:        "sha256:typed-chain-observation",
	})
	if first.Status != OriginChainStatusComplete || !first.Comparable() {
		t.Fatalf("status=%s comparable=%t missing=%v reason=%q", first.Status, first.Comparable(), first.Missing, first.Reason)
	}
	if first.Digest != second.Digest {
		t.Fatalf("equivalent origin chains must have stable digests: %q != %q", first.Digest, second.Digest)
	}
	if got := CompareOriginChains(first, second); got != OriginChainTransitionUnchanged {
		t.Fatalf("equivalent origin transition=%s, want %s", got, OriginChainTransitionUnchanged)
	}
}

func TestObserveOriginChainDoesNotPromotePartialEvidence(t *testing.T) {
	partial := ObserveOriginChain(OriginChain{
		DeclarationURI:    "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol: "activity TypedChain",
		IRNode:            "activity:TypedChain",
		MetricName:        "gooo.provenance.origin-chain.v1",
	})
	complete := ObserveOriginChain(OriginChain{
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
	if partial.Status != OriginChainStatusPartial || partial.Comparable() {
		t.Fatalf("partial status=%s comparable=%t missing=%v", partial.Status, partial.Comparable(), partial.Missing)
	}
	if got := CompareOriginChains(partial, complete); got != OriginChainTransitionUnknown {
		t.Fatalf("partial-to-complete transition=%s, want %s", got, OriginChainTransitionUnknown)
	}
	if got := CompareOriginChains(complete, ObserveOriginChain(OriginChain{
		DeclarationURI:        "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol:     "activity TypedChain",
		IRNode:                "activity:TypedChain",
		GeneratedURI:          "cmd/gooo/run_source_typed_chain_part01_test.go",
		GeneratedSymbol:       "TestTypedChain",
		ReverseObservationURI: "internal/valueexecution/replay.go",
		ReverseObservation:    "execution.replay.completed",
		MetricName:            "gooo.provenance.origin-chain.v1",
		MetricValue:           "complete=1",
		EvidenceDigest:        "sha256:changed-observation",
	})); got != OriginChainTransitionChanged {
		t.Fatalf("complete evidence change=%s, want %s", got, OriginChainTransitionChanged)
	}
}

func TestObserveOriginChainReportsUnknownWithoutEvidence(t *testing.T) {
	observation := ObserveOriginChain(OriginChain{})
	if observation.Status != OriginChainStatusUnknown || observation.Comparable() {
		t.Fatalf("empty status=%s comparable=%t reason=%q", observation.Status, observation.Comparable(), observation.Reason)
	}
	if got := CompareOriginChains(observation, observation); got != OriginChainTransitionUnknown {
		t.Fatalf("empty transition=%s, want %s", got, OriginChainTransitionUnknown)
	}
}
