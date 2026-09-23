package generation

import (
	"strings"
	"testing"
)

func TestWorkloadAuthorityProvenanceBindsIndependentIdentityDigest(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	provenance, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	expected := envelopeDigestString(strings.Join([]string{
		"spiffe-workload-identity/v1",
		"spiffe://example.org/service/billing",
	}, "\t"))
	if provenance.IdentityDigest != expected || provenance.IdentityDigest == provenance.AuthorityObservationDigest {
		t.Fatalf("identity digest = %q, authority digest = %q", provenance.IdentityDigest, provenance.AuthorityObservationDigest)
	}
	provenance.IdentityDigest = "tampered"
	if err := provenance.Validate(boundary); err == nil {
		t.Fatal("tampered identity digest was accepted")
	}
}
