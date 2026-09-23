package generation

import "testing"

func TestFoldCapabilityProvenanceDeltasPreservesOrderedHistory(t *testing.T) {
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
	series, err := FoldCapabilityProvenanceDeltas([]CapabilityProvenanceDelta{first, second})
	if err != nil {
		t.Fatalf("fold deltas: %v", err)
	}
	if len(series.Deltas) != 2 ||
		series.FirstChainDigest != first.PreviousChainDigest ||
		series.LastChainDigest != second.CurrentChainDigest ||
		series.Verdict != CapabilityDeltaChanged {
		t.Fatalf("series = %#v, want ordered two-delta history", series)
	}
	if err := series.Validate(); err != nil {
		t.Fatalf("validate series: %v", err)
	}
	if err := series.FinalDelta.Validate(start, end); err != nil {
		t.Fatalf("validate final delta: %v", err)
	}
}

func TestFoldCapabilityProvenanceDeltasPreservesUnknown(t *testing.T) {
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
	series, err := FoldCapabilityProvenanceDeltas([]CapabilityProvenanceDelta{first, second})
	if err != nil {
		t.Fatalf("fold unknown history: %v", err)
	}
	if series.Verdict != CapabilityDeltaUnknown ||
		series.FinalDelta.Verdict != CapabilityDeltaUnknown ||
		len(series.FinalDelta.Changes) != 0 {
		t.Fatalf("unknown series = %#v, want preserved UNKNOWN", series)
	}
}

func TestFoldCapabilityProvenanceDeltasRejectsEmptyAndTampering(t *testing.T) {
	if _, err := FoldCapabilityProvenanceDeltas(nil); err == nil {
		t.Fatal("empty delta series unexpectedly folded")
	}
	complete := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	delta, err := CompareCapabilityProvenanceLineage(complete, complete)
	if err != nil {
		t.Fatalf("compare complete delta: %v", err)
	}
	series, err := FoldCapabilityProvenanceDeltas([]CapabilityProvenanceDelta{delta})
	if err != nil {
		t.Fatalf("fold complete history: %v", err)
	}
	series.SeriesDigest = "tampered"
	if err := series.Validate(); err == nil {
		t.Fatal("tampered delta series unexpectedly validated")
	}
}
