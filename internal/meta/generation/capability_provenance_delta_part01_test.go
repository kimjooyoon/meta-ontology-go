package generation

import "testing"

func TestCompareCapabilityProvenanceLineageComputesReverseDelta(t *testing.T) {
	before := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	after := before
	after.Steps = append([]CapabilityProvenanceLineageStep(nil), before.Steps...)
	after.Steps[0].Digest = "source-digest-2"
	after.Steps[2].Digest = "workload-digest-2"
	after.WorkloadProvenanceDigest = "workload-digest-2"
	after.ChainDigest = after.StableHash()

	delta, err := CompareCapabilityProvenanceLineage(before, after)
	if err != nil {
		t.Fatalf("compare lineage: %v", err)
	}
	if delta.Verdict != CapabilityDeltaChanged || len(delta.Changes) != 2 {
		t.Fatalf("delta = %#v, want two changes", delta)
	}
	if err := delta.Validate(before, after); err != nil {
		t.Fatalf("validate delta: %v", err)
	}
	reversed := delta.ReverseChanges()
	if len(reversed) != 2 || reversed[0].Sequence != 2 || reversed[1].Sequence != 0 {
		t.Fatalf("reverse delta = %#v", reversed)
	}
}

func TestCompareCapabilityProvenanceLineagePreservesUnknownAndRejected(t *testing.T) {
	for _, verdict := range []string{CapabilityLineageUnknown, CapabilityLineageRejected} {
		before := testCapabilityLifecycleLineage(verdict)
		after := testCapabilityLifecycleLineage(CapabilityLineageComplete)
		delta, err := CompareCapabilityProvenanceLineage(before, after)
		if err != nil {
			t.Fatalf("%s compare lineage: %v", verdict, err)
		}
		if delta.Verdict != verdict || len(delta.Changes) != 0 {
			t.Fatalf("%s delta = %#v, want preserved terminal state without inferred changes", verdict, delta)
		}
		if err := delta.Validate(before, after); err != nil {
			t.Fatalf("%s validate delta: %v", verdict, err)
		}
	}
}

func TestCapabilityProvenanceDeltaRejectsTampering(t *testing.T) {
	before := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	after := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	delta, err := CompareCapabilityProvenanceLineage(before, after)
	if err != nil {
		t.Fatalf("compare unchanged lineage: %v", err)
	}
	delta.DeltaDigest = "tampered"
	if err := delta.Validate(before, after); err == nil {
		t.Fatal("tampered delta unexpectedly validated")
	}
}
