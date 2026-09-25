package generation

import "testing"

func cloneCapabilityLineageForComposition(source CapabilityProvenanceLineage) CapabilityProvenanceLineage {
	source.Steps = append([]CapabilityProvenanceLineageStep(nil), source.Steps...)
	return source
}

func TestComposeCapabilityProvenanceDeltaPreservesAdjacentBoundary(t *testing.T) {
	start := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	middle := cloneCapabilityLineageForComposition(start)
	middle.Steps[0].Digest = "source-digest-middle"
	middle.ChainDigest = middle.StableHash()

	end := cloneCapabilityLineageForComposition(middle)
	end.Steps[2].Digest = "workload-digest-end"
	end.WorkloadProvenanceDigest = "workload-digest-end"
	end.ChainDigest = end.StableHash()

	first, err := CompareCapabilityProvenanceLineage(start, middle)
	if err != nil {
		t.Fatalf("compare first delta: %v", err)
	}
	second, err := CompareCapabilityProvenanceLineage(middle, end)
	if err != nil {
		t.Fatalf("compare second delta: %v", err)
	}
	composed, err := ComposeCapabilityProvenanceDelta(first, second)
	if err != nil {
		t.Fatalf("compose deltas: %v", err)
	}
	if composed.Verdict != CapabilityDeltaChanged || len(composed.Changes) != 2 {
		t.Fatalf("composed delta = %#v, want two changes", composed)
	}
	if composed.Changes[0].BeforeDigest != start.Steps[0].Digest ||
		composed.Changes[0].AfterDigest != middle.Steps[0].Digest {
		t.Fatalf("source composition = %#v", composed.Changes[0])
	}
	if composed.Changes[1].BeforeDigest != middle.Steps[2].Digest ||
		composed.Changes[1].AfterDigest != end.Steps[2].Digest {
		t.Fatalf("workload composition = %#v", composed.Changes[1])
	}
	if err := composed.Validate(start, end); err != nil {
		t.Fatalf("validate composed delta: %v", err)
	}
	reversed := composed.ReverseChanges()
	if len(reversed) != 2 || reversed[0].Sequence != 2 || reversed[1].Sequence != 0 {
		t.Fatalf("reverse composed delta = %#v", reversed)
	}
}

func TestComposeCapabilityProvenanceDeltaPreservesTerminalVerdict(t *testing.T) {
	unknown := testCapabilityLifecycleLineage(CapabilityLineageUnknown)
	complete := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	first, err := CompareCapabilityProvenanceLineage(unknown, complete)
	if err != nil {
		t.Fatalf("compare unknown delta: %v", err)
	}
	second, err := CompareCapabilityProvenanceLineage(complete, complete)
	if err != nil {
		t.Fatalf("compare complete delta: %v", err)
	}
	composed, err := ComposeCapabilityProvenanceDelta(first, second)
	if err != nil {
		t.Fatalf("compose terminal delta: %v", err)
	}
	if composed.Verdict != CapabilityDeltaUnknown || len(composed.Changes) != 0 {
		t.Fatalf("terminal composed delta = %#v, want preserved UNKNOWN", composed)
	}
}

func TestComposeCapabilityProvenanceDeltaRejectsBoundaryAndDigestTampering(t *testing.T) {
	start := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	middle := cloneCapabilityLineageForComposition(start)
	middle.Steps[0].Digest = "source-digest-middle"
	middle.ChainDigest = middle.StableHash()
	first, err := CompareCapabilityProvenanceLineage(start, middle)
	if err != nil {
		t.Fatalf("compare first delta: %v", err)
	}
	second, err := CompareCapabilityProvenanceLineage(middle, middle)
	if err != nil {
		t.Fatalf("compare second delta: %v", err)
	}

	tampered := second
	tampered.DeltaDigest = "tampered"
	if _, err := ComposeCapabilityProvenanceDelta(first, tampered); err == nil {
		t.Fatal("tampered delta unexpectedly composed")
	}

	mismatched := second
	mismatched.PreviousChainDigest = "wrong-boundary"
	mismatched.DeltaDigest = mismatched.StableHash()
	if _, err := ComposeCapabilityProvenanceDelta(first, mismatched); err == nil {
		t.Fatal("mismatched boundary unexpectedly composed")
	}
}
