package generation

import "testing"

func TestObserveWorkloadAuthorityProvenanceBindsIdentityToBoundary(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	provenance, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	if provenance.BindingStatus != WorkloadBindingBound || provenance.TrustDomain != "example.org" || provenance.WorkloadPath != "/service/billing" || !provenance.NonAuthorizing {
		t.Fatalf("unexpected workload provenance: %#v", provenance)
	}
	if provenance.AuthorityObservationDigest == "" || provenance.StableHash() == "" {
		t.Fatal("workload provenance did not preserve observation digests")
	}
	if err := provenance.Validate(boundary); err != nil {
		t.Fatalf("validate workload provenance: %v", err)
	}
}

func TestObserveWorkloadAuthorityProvenancePreservesBoundaryUncertainty(t *testing.T) {
	stale, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r2"))
	if err != nil {
		t.Fatal(err)
	}
	provenance, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", stale)
	if err != nil {
		t.Fatal(err)
	}
	if provenance.BindingStatus != WorkloadBindingRejected || provenance.BoundaryStatus != "STALE_RESULT" {
		t.Fatalf("stale boundary was not rejected: %#v", provenance)
	}

	missingResult := authorityBoundaryIR([]string{"read:source"}, "source-r1")
	missingResult.Result = nil
	missingResultObservation, err := ObserveAuthorityBoundary(missingResult)
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", missingResultObservation)
	if err != nil {
		t.Fatal(err)
	}
	if unknown.BindingStatus != WorkloadBindingUnknown {
		t.Fatalf("unobserved result was promoted: %#v", unknown)
	}
}

func TestObserveWorkloadAuthorityProvenanceRejectsInvalidIdentityAndTampering(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ObserveWorkloadAuthorityProvenance("https://example.org/service/billing", boundary); err == nil {
		t.Fatal("non-SPIFFE identity was accepted")
	}
	provenance, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	provenance.AuthorityObservationDigest = "tampered"
	if err := provenance.Validate(boundary); err == nil {
		t.Fatal("tampered workload provenance was accepted")
	}
}
