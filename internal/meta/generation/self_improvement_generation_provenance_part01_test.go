package generation

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementGenerationProvenanceBindsAllEvidence(t *testing.T) {
	digest := strings.Repeat("a", 64)
	observation := ObserveSelfImprovementGenerationProvenance(SelfImprovementGenerationProvenanceInput{
		DeclarationDigest:        digest,
		SourceDigest:             strings.Repeat("b", 64),
		SemanticDigest:           strings.Repeat("c", 64),
		GraphDigest:              strings.Repeat("d", 64),
		GeneratedDigest:          strings.Repeat("e", 64),
		ReverseObservationDigest: strings.Repeat("f", 64),
	})
	if observation.BindingStatus != SelfImprovementGenerationProvenanceBound {
		t.Fatalf("binding status = %q, want BOUND", observation.BindingStatus)
	}
	if observation.Reason != "GENERATION_EVIDENCE_BOUND" {
		t.Fatalf("reason = %q, want complete evidence", observation.Reason)
	}
	if !observation.NonAuthorizing {
		t.Fatal("generation provenance must remain non-authorizing")
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("validate observation: %v", err)
	}
}

func TestObserveSelfImprovementGenerationProvenancePreservesUnknownCause(t *testing.T) {
	observation := ObserveSelfImprovementGenerationProvenance(SelfImprovementGenerationProvenanceInput{
		DeclarationDigest: strings.Repeat("a", 64),
		SourceDigest:      strings.Repeat("b", 64),
		SemanticDigest:    strings.Repeat("c", 64),
		GraphDigest:       strings.Repeat("d", 64),
	})
	if observation.BindingStatus != SelfImprovementGenerationProvenanceUnknown {
		t.Fatalf("binding status = %q, want UNKNOWN", observation.BindingStatus)
	}
	if observation.Reason != "GENERATED_DIGEST_MISSING" {
		t.Fatalf("reason = %q, want generated digest cause", observation.Reason)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("validate unknown observation: %v", err)
	}
}

func TestSelfImprovementGenerationProvenanceDigestTracksReverseObservation(t *testing.T) {
	base := SelfImprovementGenerationProvenance{
		Schema:                   SelfImprovementGenerationProvenanceSchema,
		DeclarationDigest:        strings.Repeat("a", 64),
		SourceDigest:             strings.Repeat("b", 64),
		SemanticDigest:           strings.Repeat("c", 64),
		GraphDigest:              strings.Repeat("d", 64),
		GeneratedDigest:          strings.Repeat("e", 64),
		ReverseObservationDigest: strings.Repeat("f", 64),
		BindingStatus:            SelfImprovementGenerationProvenanceBound,
		Reason:                   "GENERATION_EVIDENCE_BOUND",
		NonAuthorizing:           true,
	}
	changed := base
	changed.ReverseObservationDigest = strings.Repeat("0", 64)
	if base.StableHash() == changed.StableHash() {
		t.Fatal("reverse observation digest must affect provenance identity")
	}
}