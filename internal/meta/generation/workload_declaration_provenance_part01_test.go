package generation

import "testing"

func TestWorkloadDeclarationProvenanceBindsDeclarationToWorkload(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	workload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	declaration := envelopeDigestString("gooo-declaration/examples/billing.gooo")
	observation, err := ObserveWorkloadDeclarationProvenance(workload, declaration)
	if err != nil {
		t.Fatal(err)
	}
	if observation.BindingStatus != WorkloadDeclarationBindingBound ||
		observation.DeclarationDigest != declaration ||
		!observation.NonAuthorizing {
		t.Fatalf("unexpected declaration provenance: %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("validate declaration provenance: %v", err)
	}
}

func TestWorkloadDeclarationProvenancePreservesRejectedBinding(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r2"))
	if err != nil {
		t.Fatal(err)
	}
	workload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ObserveWorkloadDeclarationProvenance(workload, "not-a-digest")
	if err == nil {
		t.Fatal("malformed declaration digest was accepted")
	}
}
