package semantic

import (
	"strings"
	"testing"
)

func inferenceObservationChainFixture() InferencePathChain {
	declaration := inferenceEdgeFixture(InferenceAuthoritativeDeclaration, "binding-declaration")
	derivation := inferenceEdgeFixture(InferenceDeterministicDerivation, "binding-derivation")
	observation := inferenceEdgeFixture(InferenceObservationCandidate, "binding-observation")
	derivation.SubjectID = declaration.ObjectID
	observation.SubjectID = derivation.ObjectID
	return InferencePathChain{Edges: []InferenceEdge{declaration, derivation, observation}}
}

func TestBindInferenceObservationComputesDeclarationIRAndObservationIdentity(t *testing.T) {
	chain := inferenceObservationChainFixture()
	binding, err := BindInferenceObservation(chain)
	if err != nil {
		t.Fatalf("bind inference observation: %v", err)
	}
	if binding.Schema != InferenceObservationBindingSchema || binding.DeclarationID != chain.Edges[0].SubjectID || binding.SemanticID != chain.Edges[0].ObjectID || binding.ObservationID != chain.Edges[2].ObjectID {
		t.Fatalf("unexpected binding identity: %#v", binding)
	}
	if binding.SourceDigest != chain.Edges[0].After.Source || binding.SemanticDigest != chain.Edges[2].After.Semantic || binding.ChainDigest == "" || binding.EvidenceDigest == "" {
		t.Fatalf("binding did not retain exact snapshots: %#v", binding)
	}
	if err := binding.Validate(chain); err != nil {
		t.Fatalf("validate binding: %v", err)
	}
	if binding.StableHash() != StableHashString(binding.Canonical()) {
		t.Fatal("binding stable hash is not canonical")
	}
}

func TestInferenceObservationBindingRejectsIncompleteOrTamperedPaths(t *testing.T) {
	chain := inferenceObservationChainFixture()
	if _, err := BindInferenceObservation(InferencePathChain{Edges: chain.Edges[:2]}); err == nil {
		t.Fatal("incomplete path was accepted")
	}
	withoutDerivation := inferenceObservationChainFixture()
	withoutDerivation.Edges[1].Kind = InferenceObservationCandidate
	if _, err := BindInferenceObservation(withoutDerivation); err == nil {
		t.Fatal("path without derivation was accepted")
	}
	binding, err := BindInferenceObservation(chain)
	if err != nil {
		t.Fatal(err)
	}
	binding.SemanticDigest = strings.Repeat("0", 64)
	if err := binding.Validate(chain); err == nil {
		t.Fatal("tampered semantic digest was accepted")
	}
}
