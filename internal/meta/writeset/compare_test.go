package writeset

import "testing"

func TestCompareExactAndMismatch(t *testing.T) {
	before := testSnapshot(Entry{Path: "value.txt", Kind: "FILE", Size: 1, Digest: "sha256:before"})
	after := testSnapshot(Entry{Path: "value.txt", Kind: "FILE", Size: 1, Digest: "sha256:after"})
	exact := Compare("subject", "denominator", before, after, []string{"value.txt"})
	if exact.Decision != "PASS" || exact.Summary.ExactnessBPS != 10000 || exact.Summary.MismatchPaths != 0 { t.Fatalf("unexpected exact receipt: %+v", exact) }
	mismatch := Compare("subject", "denominator", before, after, nil)
	if mismatch.Decision != "BLOCK" || mismatch.Summary.ExactnessBPS != 0 || mismatch.Summary.MismatchPaths != 1 { t.Fatalf("unexpected mismatch receipt: %+v", mismatch) }
}

func TestCompareUnknownFailsClosed(t *testing.T) {
	receipt := Compare("subject", "denominator", Snapshot{}, Snapshot{}, nil)
	if receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "WRITE_SET_EVIDENCE_UNKNOWN" || receipt.Resolution != "INVARIANT_ONLY" { t.Fatalf("unexpected unknown receipt: %+v", receipt) }
}

func testSnapshot(entries ...Entry) Snapshot { return Snapshot{Schema: SnapshotSchema, RootDigest: digestEntries(entries), Entries: entries} }
