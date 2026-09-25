package lanefrontier

import (
	"bytes"
	"strings"
	"testing"
)

func TestCorpus(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if corpus.Schema != "gooo/lane-frontier/corpus/v1" {
		t.Fatalf("corpus schema = %q", corpus.Schema)
	}
	reasons := map[Reason]bool{UnknownSchema: false, MissingInput: false, InvalidCount: false, AmbiguousOwner: false, PathOutOfScope: false, ActiveLease: false, ActivePR: false, DivergedBranch: false, StaleBranch: false, BranchAhead: false, Clean: false}
	names := map[string]bool{}
	for _, fixture := range corpus.Cases {
		if names[fixture.Name] {
			t.Fatalf("duplicate corpus case %q", fixture.Name)
		}
		names[fixture.Name] = true
		got := Evaluate(fixture.Input)
		if got.Decision != fixture.Expected.Decision || got.Reason != fixture.Expected.Reason {
			t.Errorf("%s result = %#v, want %#v", fixture.Name, got, fixture.Expected)
		}
		if got.CanonicalDigest != fixture.CanonicalDigest || got.CanonicalDigest != fixture.Expected.CanonicalDigest || fixture.CanonicalDigest != CanonicalDigest(fixture) {
			t.Errorf("%s digest = %q, want case=%q expected=%q canonical=%q", fixture.Name, got.CanonicalDigest, fixture.CanonicalDigest, fixture.Expected.CanonicalDigest, CanonicalDigest(fixture))
		}
		if strings.Contains(string(got.Decision), "N/A") || strings.Contains(string(got.Decision), "PASS") {
			t.Errorf("%s has forbidden decision %q", fixture.Name, got.Decision)
		}
		if _, ok := reasons[fixture.Expected.Reason]; !ok {
			t.Errorf("%s has unlisted reason %q", fixture.Name, fixture.Expected.Reason)
		} else {
			reasons[fixture.Expected.Reason] = true
		}
	}
	for reason, present := range reasons {
		if !present {
			t.Errorf("corpus omitted reason %q", reason)
		}
	}
	t.Logf("lane frontier corpus cases=%d corpus_sha=%s", len(corpus.Cases), CorpusDigest())
}
func TestCanonicalPermutation(t *testing.T) {
	base := permutationCase()
	permuted := base
	permuted.Input.OwnedPathPrefixes = []string{"pkg/z", "pkg/a"}
	permuted.Input.ChangedPaths = []string{"pkg/z/file.go", "pkg/a/file.go"}
	if !bytes.Equal(CanonicalJSON(base), CanonicalJSON(permuted)) {
		t.Fatalf("canonical bytes changed under input permutation")
	}
	if Evaluate(base.Input) != Evaluate(permuted.Input) {
		t.Fatalf("decision changed under input permutation")
	}
}
func TestStrictJSONRejectsDuplicateAndUnknownFields(t *testing.T) {
	if _, err := DecodeInput([]byte(`{"schema":"gooo/lane-frontier/v1","schema":"gooo/lane-frontier/v1"}`)); err == nil {
		t.Fatal("duplicate field accepted")
	}
	if _, err := DecodeInput([]byte(`{"schema":"gooo/lane-frontier/v1","unknown":1}`)); err == nil {
		t.Fatal("unknown field accepted")
	}
}
