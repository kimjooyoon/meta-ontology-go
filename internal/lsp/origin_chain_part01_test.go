package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestNewOriginChainReferencePreservesComparableEvidence(t *testing.T) {
	observation := provenance.ObserveOriginChain(provenance.OriginChain{
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

	reference, comparable := NewOriginChainReference(observation)
	if !comparable || reference.Status != provenance.OriginChainStatusComplete {
		t.Fatalf("comparable=%t status=%s reason=%q", comparable, reference.Status, reference.Reason)
	}
	if reference.Digest != observation.Digest || reference.EvidenceDigest != observation.Chain.EvidenceDigest {
		t.Fatalf("reference lost evidence identity: %#v", reference)
	}
	if !reference.NonAuthorizing {
		t.Fatal("LSP origin reference must remain non-authorizing")
	}
}

func TestNewOriginChainReferenceDoesNotPromoteUnknownEvidence(t *testing.T) {
	observation := provenance.ObserveOriginChain(provenance.OriginChain{
		DeclarationURI:    "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol: "activity TypedChain",
	})

	reference, comparable := NewOriginChainReference(observation)
	if comparable || reference.Status != provenance.OriginChainStatusPartial {
		t.Fatalf("comparable=%t status=%s missing reason=%q", comparable, reference.Status, reference.Reason)
	}
	if reference.NonAuthorizing != true || reference.Reason == "" {
		t.Fatalf("incomplete reference lost its guardrails: %#v", reference)
	}
}
