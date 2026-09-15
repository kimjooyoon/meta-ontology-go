package resourcevector

import (
	"testing"
)

func TestR4F01Vectors(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if corpus.Schema != CorpusSchemaV1 || len(corpus.Cases) != 1 {
		t.Fatalf("corpus = %q/%d", corpus.Schema, len(corpus.Cases))
	}
	if corpus.CanonicalDigest != CorpusDigest(corpus) {
		t.Fatalf("corpus digest = %q, want %q", corpus.CanonicalDigest, CorpusDigest(corpus))
	}
	row := corpus.Cases[0]
	got := Evaluate(row.Input)
	if got.Decision != DecisionPass || got.Reason != ReasonNone || !got.ProofValid {
		t.Fatalf("r4-f-01 output = %#v", got)
	}
	wantSelected := Vector{CPUCoreNS: 33, MemoryBytes: 224, PeakMemoryBytes: 128, WorkUnits: 18, AffectedStableIDs: 2, ApplicablePressures: 2, IndependentGroups: 2, UniquePROVRecords: 12, FinitePROVPaths: 3, ClosureNumerator: 3, ClosureDenominator: 3}
	wantFull := Vector{CPUCoreNS: 36, MemoryBytes: 272, PeakMemoryBytes: 128, WorkUnits: 19, AffectedStableIDs: 2, ApplicablePressures: 3, IndependentGroups: 3, UniquePROVRecords: 15, FinitePROVPaths: 4, ClosureNumerator: 4, ClosureDenominator: 4}
	if got.Selected == nil || got.Full == nil || *got.Selected != wantSelected || *got.Full != wantFull {
		t.Fatalf("vectors selected=%#v full=%#v", got.Selected, got.Full)
	}
	if row.Expected.Selected == nil || row.Expected.Full == nil || *row.Expected.Selected != wantSelected || *row.Expected.Full != wantFull {
		t.Fatalf("fixture expected vectors selected=%#v full=%#v", row.Expected.Selected, row.Expected.Full)
	}
	if got.InputDigest != row.Expected.InputDigest || got.CanonicalOutputDigest != row.Expected.CanonicalOutputDigest || got.ReplayDigest != row.Expected.ReplayDigest {
		t.Fatalf("fixture digests got=%#v expected=%#v", got, row.Expected)
	}
	t.Logf("input=%s output=%s replay=%s corpus=%s", got.InputDigest, got.CanonicalOutputDigest, got.ReplayDigest, CorpusDigest(corpus))
}
