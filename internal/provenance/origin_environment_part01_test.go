package provenance

import (
	"strings"
	"testing"
)

func completeOriginEnvironmentObservation() OriginChainObservation {
	return ObserveOriginChain(OriginChain{
		DeclarationURI:        "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol:     "typed_chain",
		IRNode:                "ir:typed_chain",
		GeneratedURI:          "generated://typed_chain",
		GeneratedSymbol:       "typed_chain",
		ReverseObservationURI: "observation://typed_chain",
		ReverseObservation:    "execution_digest",
		MetricName:            "replay_state",
		MetricValue:           "CLOSED",
		EvidenceDigest:        strings.Repeat("a", 64),
	})
}

func TestObserveOriginEnvironmentSeparatesOriginAndEnvironment(t *testing.T) {
	origin := completeOriginEnvironmentObservation()
	first := ObserveOriginEnvironment(origin, "sha256:"+strings.Repeat("b", 64))
	if first.Status != OriginEnvironmentStatusBound || !first.NonAuthorizing || first.EvidenceDigest == "" {
		t.Fatalf("expected bound non-authorizing evidence, got %#v", first)
	}
	if got := CompareOriginEnvironments(first, first); got != OriginEnvironmentTransitionUnchanged {
		t.Fatalf("same environment transition = %q", got)
	}
	second := ObserveOriginEnvironment(origin, "sha256:"+strings.Repeat("c", 64))
	if got := CompareOriginEnvironments(first, second); got != OriginEnvironmentTransitionChanged {
		t.Fatalf("changed environment transition = %q", got)
	}
}

func TestObserveOriginEnvironmentPreservesIncompleteEvidence(t *testing.T) {
	unknown := ObserveOriginEnvironment(OriginChainObservation{}, "sha256:"+strings.Repeat("b", 64))
	if unknown.Status != OriginEnvironmentStatusUnknown || unknown.Reason != "ORIGIN_OBSERVATION_INCOMPLETE" {
		t.Fatalf("expected incomplete origin to remain unknown, got %#v", unknown)
	}
	invalid := ObserveOriginEnvironment(completeOriginEnvironmentObservation(), "not-a-digest")
	if invalid.Status != OriginEnvironmentStatusUnknown || invalid.Reason != "ENVIRONMENT_DIGEST_INVALID" {
		t.Fatalf("expected invalid environment to remain unknown, got %#v", invalid)
	}
}
