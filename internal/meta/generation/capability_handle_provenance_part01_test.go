package generation

import "testing"

func TestObserveCapabilityHandleProvenanceBindsGrantAndWorkload(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	workload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	provenance, err := ObserveCapabilityHandleProvenance(workload, boundary, "handle-1", "gooo://capability/source", "read:source")
	if err != nil {
		t.Fatal(err)
	}
	if provenance.ObservationStatus != CapabilityHandleObserved || !provenance.EffectWithinGrant || !provenance.NonAuthorizing {
		t.Fatalf("unexpected capability handle provenance: %#v", provenance)
	}
	if err := provenance.Validate(workload, boundary); err != nil {
		t.Fatalf("validate capability handle provenance: %v", err)
	}
}

func TestObserveCapabilityHandleProvenancePreservesUnknownAndRejection(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"write:repository"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	workload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	rejected, err := ObserveCapabilityHandleProvenance(workload, boundary, "handle-2", "gooo://capability/source", "write:repository")
	if err != nil {
		t.Fatal(err)
	}
	if rejected.ObservationStatus != CapabilityHandleRejected || rejected.EffectWithinGrant {
		t.Fatalf("escalated capability handle was not rejected: %#v", rejected)
	}

	missingResult := authorityBoundaryIR([]string{"read:source"}, "source-r1")
	missingResult.Result = nil
	unknownBoundary, err := ObserveAuthorityBoundary(missingResult)
	if err != nil {
		t.Fatal(err)
	}
	unknownWorkload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", unknownBoundary)
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := ObserveCapabilityHandleProvenance(unknownWorkload, unknownBoundary, "handle-3", "gooo://capability/source", "read:source")
	if err != nil {
		t.Fatal(err)
	}
	if unknown.ObservationStatus != CapabilityHandleUnknown {
		t.Fatalf("missing result was promoted: %#v", unknown)
	}
}

func TestObserveCapabilityHandleProvenanceRejectsInvalidURIAndTampering(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	workload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ObserveCapabilityHandleProvenance(workload, boundary, "handle-4", "https://example.org/source", "read:source"); err == nil {
		t.Fatal("non-gooo capability URI was accepted")
	}
	provenance, err := ObserveCapabilityHandleProvenance(workload, boundary, "handle-5", "gooo://capability/source", "read:source")
	if err != nil {
		t.Fatal(err)
	}
	provenance.Effect = "tampered"
	if err := provenance.Validate(workload, boundary); err == nil {
		t.Fatal("tampered capability handle provenance was accepted")
	}
}
