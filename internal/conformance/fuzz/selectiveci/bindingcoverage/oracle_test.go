package bindingcoverage

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCorpus(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if corpus.Schema != SchemaV1 {
		t.Fatalf("schema=%q", corpus.Schema)
	}
	if len(corpus.Cases) != 20 {
		t.Fatalf("case count=%d", len(corpus.Cases))
	}
	digest, err := CorpusDigest(corpus)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("binding coverage corpus cases=%d corpus_sha=%s", len(corpus.Cases), digest)
	if corpus.CanonicalDigest != digest {
		t.Errorf("canonical corpus digest: want %q, got %q", corpus.CanonicalDigest, digest)
	}
	seen := make(map[string]bool, len(corpus.Cases))
	for _, row := range corpus.Cases {
		if row.Name == "" || seen[row.Name] {
			t.Errorf("duplicate or empty case name %q", row.Name)
		}
		seen[row.Name] = true
		got := Evaluate(row.Input)
		if !reflect.DeepEqual(row.Expected, got.Vector) || row.ExpectedDigest != got.CanonicalDigest {
			actual, _ := json.Marshal(got.Vector)
			t.Errorf("case %s vector mismatch: actual=%s digest=%s", row.Name, actual, got.CanonicalDigest)
		}
		if row.ExpectedDigest != digestVector(row.Expected) {
			t.Errorf("case %s expected digest does not seal expected vector", row.Name)
		}
	}
}

func TestRequiredCoverageShapes(t *testing.T) {
	corpus := mustCorpus(t)
	rows := indexCases(corpus)
	complete := Evaluate(rows["selective-ci-9-binding-complete"].Input)
	if complete.Decision != DecisionExact || complete.RequiredBindingCount < 9 {
		t.Fatalf("nine-binding shape=%+v", complete.Vector)
	}
	missing := Evaluate(rows["selective-ci-9-missing-lane-registry-mismatch"].Input)
	if missing.Decision != DecisionIncomplete || !reflect.DeepEqual(missing.MissingMismatch, []string{"sid:lane-registry"}) || len(missing.MissingMatch) != 0 {
		t.Fatalf("lane registry omission=%+v", missing.Vector)
	}
	if missing.ExecutionAuthorized || missing.CIAuthorized {
		t.Fatal("authorization escaped the oracle partition")
	}
	shared := Evaluate(rows["shared-endpoint-references"].Input)
	if shared.Decision != DecisionExact || shared.EndpointReferenceCount != 4 || shared.WorkUnits != 10 {
		t.Fatalf("shared endpoint references=%+v", shared.Vector)
	}
	stale := Evaluate(rows["valid-but-unequal-snapshot-digests"].Input)
	if stale.Decision != DecisionUnknown || stale.Reason != "STALE_OR_BAD_DIGEST" {
		t.Fatalf("unequal snapshot digests=%+v", stale.Vector)
	}
	selfLink := Evaluate(rows["self-link-binding"].Input)
	if selfLink.Decision != DecisionUnknown || selfLink.Reason != "UNKNOWN_BINDING" {
		t.Fatalf("self-link binding=%+v", selfLink.Vector)
	}
}

func TestCanonicalPermutation(t *testing.T) {
	corpus := mustCorpus(t)
	rows := indexCases(corpus)
	left := Evaluate(rows["complete-two-bindings"].Input)
	right := Evaluate(rows["permuted-complete-two-bindings"].Input)
	if !reflect.DeepEqual(left.Vector, right.Vector) || left.CanonicalDigest != right.CanonicalDigest {
		t.Fatalf("permutation changed result: left=%+v right=%+v", left, right)
	}
}

func TestStrictJSONRejectsDuplicatesAndUnknownFields(t *testing.T) {
	var input Input
	if err := decodeStrict([]byte(`{"schema":"binding-coverage/v1","schema":"binding-coverage/v1"}`), &input); err == nil {
		t.Fatal("duplicate JSON field accepted")
	}
	if err := decodeStrict([]byte(`{"schema":"binding-coverage/v1","unknown":1}`), &input); err == nil {
		t.Fatal("unknown JSON field accepted")
	}
}

func TestIntegerOverflowBoundaryRejectsOverflow(t *testing.T) {
	var input Input
	data := []byte(`{"schema":"binding-coverage/v1","snapshot_digest":"sha256:0000000000000000000000000000000000000000000000000000000000000000","expected_snapshot_digest":"sha256:0000000000000000000000000000000000000000000000000000000000000000","precedence":[{"rank":9223372036854775808,"stage":"stage:binding","reason":"reason:evidence"}],"required_bindings":[],"partitions":[]}`)
	if err := decodeStrict(data, &input); err == nil {
		t.Fatal("integer overflow accepted")
	}
}

func FuzzEvaluateNeverPanics(f *testing.F) {
	corpus := mustCorpus(f)
	for _, row := range corpus.Cases {
		data, err := canonicalInputJSON(row.Input)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var input Input
		if decodeStrict(data, &input) == nil {
			_ = Evaluate(input)
		}
	})
}

func mustCorpus(t testing.TB) CorpusFile {
	t.Helper()
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	return corpus
}

func indexCases(corpus CorpusFile) map[string]CorpusCase {
	rows := make(map[string]CorpusCase, len(corpus.Cases))
	for _, row := range corpus.Cases {
		rows[row.Name] = row
	}
	return rows
}
