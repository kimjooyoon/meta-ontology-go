package generation

import "testing"

func TestObserveCapabilityProvenanceLineageComputesAndReversesOrigin(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	workload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := ObserveCapabilityHandleProvenance(workload, boundary, "handle-1", "gooo://capability/source", "read:source")
	if err != nil {
		t.Fatal(err)
	}
	lineage, err := ObserveCapabilityProvenanceLineage(workload, boundary, capability)
	if err != nil {
		t.Fatal(err)
	}
	if lineage.Verdict != CapabilityLineageComplete || len(lineage.Steps) != 4 || !lineage.NonAuthorizing {
		t.Fatalf("unexpected capability lineage: %#v", lineage)
	}
	reversed := lineage.ReverseOrigin()
	if reversed[0].Phase != "CAPABILITY_HANDLE" || reversed[len(reversed)-1].Phase != "SOURCE_REVISION" {
		t.Fatalf("reverse origin = %#v", reversed)
	}
	if err := lineage.Validate(workload, boundary, capability); err != nil {
		t.Fatalf("validate capability lineage: %v", err)
	}
}

func TestObserveCapabilityProvenanceLineagePreservesUnknownAndRejectedVerdicts(t *testing.T) {
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
	unknownCapability, err := ObserveCapabilityHandleProvenance(unknownWorkload, unknownBoundary, "handle-2", "gooo://capability/source", "read:source")
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := ObserveCapabilityProvenanceLineage(unknownWorkload, unknownBoundary, unknownCapability)
	if err != nil {
		t.Fatal(err)
	}
	if unknown.Verdict != CapabilityLineageUnknown {
		t.Fatalf("unknown lineage was promoted: %#v", unknown)
	}

	rejectedBoundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"write:repository"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	rejectedWorkload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", rejectedBoundary)
	if err != nil {
		t.Fatal(err)
	}
	rejectedCapability, err := ObserveCapabilityHandleProvenance(rejectedWorkload, rejectedBoundary, "handle-3", "gooo://capability/source", "write:repository")
	if err != nil {
		t.Fatal(err)
	}
	rejected, err := ObserveCapabilityProvenanceLineage(rejectedWorkload, rejectedBoundary, rejectedCapability)
	if err != nil {
		t.Fatal(err)
	}
	if rejected.Verdict != CapabilityLineageRejected {
		t.Fatalf("rejected lineage was promoted: %#v", rejected)
	}
}

func TestObserveCapabilityProvenanceLineageRejectsTampering(t *testing.T) {
	boundary, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	workload, err := ObserveWorkloadAuthorityProvenance("spiffe://example.org/service/billing", boundary)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := ObserveCapabilityHandleProvenance(workload, boundary, "handle-4", "gooo://capability/source", "read:source")
	if err != nil {
		t.Fatal(err)
	}
	lineage, err := ObserveCapabilityProvenanceLineage(workload, boundary, capability)
	if err != nil {
		t.Fatal(err)
	}
	lineage.Steps[0].Digest = "tampered"
	if err := lineage.Validate(workload, boundary, capability); err == nil {
		t.Fatal("tampered capability lineage was accepted")
	}
}
