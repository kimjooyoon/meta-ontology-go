package generation

import "testing"

func TestCapabilityProvenanceLifecycleReversesCompleteLineage(t *testing.T) {
	lineage := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	lifecycle, err := ObserveCapabilityProvenanceLifecycle(lineage, []string{
		CapabilityLifecycleBorrowed,
		CapabilityLifecycleOwned,
		CapabilityLifecycleDropped,
	})
	if err != nil {
		t.Fatalf("observe lifecycle: %v", err)
	}
	if lifecycle.FinalStatus != CapabilityLifecycleDropped {
		t.Fatalf("final status = %q, want %q", lifecycle.FinalStatus, CapabilityLifecycleDropped)
	}
	if err := lifecycle.Validate(lineage); err != nil {
		t.Fatalf("validate lifecycle: %v", err)
	}
	reversed := lifecycle.ReverseLifecycle()
	if len(reversed) != 3 || reversed[0].Mode != CapabilityLifecycleDropped || reversed[2].Mode != CapabilityLifecycleBorrowed {
		t.Fatalf("reverse lifecycle = %#v", reversed)
	}
}

func TestCapabilityProvenanceLifecyclePreservesUnknownAndRejected(t *testing.T) {
	for _, verdict := range []string{CapabilityLineageUnknown, CapabilityLineageRejected} {
		lineage := testCapabilityLifecycleLineage(verdict)
		lifecycle, err := ObserveCapabilityProvenanceLifecycle(lineage, nil)
		if err != nil {
			t.Fatalf("%s observe lifecycle: %v", verdict, err)
		}
		if lifecycle.FinalStatus != verdict || len(lifecycle.Events) != 0 {
			t.Fatalf("%s lifecycle = %#v, want terminal status with no inferred events", verdict, lifecycle)
		}
		if err := lifecycle.Validate(lineage); err != nil {
			t.Fatalf("%s validate lifecycle: %v", verdict, err)
		}
	}
}

func TestCapabilityProvenanceLifecycleRejectsTampering(t *testing.T) {
	lineage := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	lifecycle, err := ObserveCapabilityProvenanceLifecycle(lineage, []string{CapabilityLifecycleBorrowed})
	if err != nil {
		t.Fatalf("observe lifecycle: %v", err)
	}
	lifecycle.LifecycleDigest = "tampered"
	if err := lifecycle.Validate(lineage); err == nil {
		t.Fatal("tampered lifecycle unexpectedly validated")
	}
}

func testCapabilityLifecycleLineage(verdict string) CapabilityProvenanceLineage {
	lineage := CapabilityProvenanceLineage{
		Schema: CapabilityProvenanceLineageSchema,
		SourceRevision: SourceRevision{
			ID:     "source-1",
			Digest: "source-digest-1",
		},
		OperationID: "operation-1",
		Steps: []CapabilityProvenanceLineageStep{
			{Phase: "SOURCE_REVISION", ID: "source-1", Digest: "source-digest-1"},
			{Phase: "AUTHORITY_BOUNDARY", ID: "operation-1", Digest: "authority-digest-1"},
			{Phase: "WORKLOAD_IDENTITY", ID: "spiffe://example.test/ns/demo/sa/gooo", Digest: "workload-digest-1"},
			{Phase: "CAPABILITY_HANDLE", ID: "handle-1", Digest: "handle-digest-1"},
		},
		AuthorityObservationDigest: "authority-observation-1",
		WorkloadProvenanceDigest:   "workload-digest-1",
		CapabilityHandleDigest:     "handle-digest-1",
		Verdict:                    verdict,
		NonAuthorizing:             true,
	}
	lineage.ChainDigest = lineage.StableHash()
	return lineage
}
