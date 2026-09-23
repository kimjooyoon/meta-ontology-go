package generation

import "testing"

func TestCapabilityProvenanceSurfaceObservationPreservesUnknownBindings(t *testing.T) {
	observation := ObserveCapabilityProvenanceSurface(CapabilityProvenanceSurfaceObservationInput{})
	if observation.Decision != CapabilityProvenanceSurfaceUnknown || observation.Reason != "MISSING_SOURCE_DIGEST" || !observation.NonAuthorizing {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("unknown observation should be structurally valid: %v", err)
	}
}

func TestCapabilityProvenanceSurfaceObservationRejectsMalformedBinding(t *testing.T) {
	observation := ObserveCapabilityProvenanceSurface(CapabilityProvenanceSurfaceObservationInput{
		SourceDigest: "not-a-digest", SemanticDigest: "sha256:" + "0", GeneratedDigest: "sha256:" + "0", ReverseObservationDigest: "sha256:" + "0",
	})
	if observation.Decision != CapabilityProvenanceSurfaceRefuted || observation.Reason != "MALFORMED_SOURCE_DIGEST" {
		t.Fatalf("observation = %#v", observation)
	}
}

func TestCapabilityProvenanceSurfaceObservationDigestBindsDecision(t *testing.T) {
	observation := ObserveCapabilityProvenanceSurface(CapabilityProvenanceSurfaceObservationInput{})
	observation.Reason = "tampered"
	if err := observation.Validate(); err == nil {
		t.Fatal("tampered surface observation was accepted")
	}
}
