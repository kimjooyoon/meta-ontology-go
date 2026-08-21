package bindingcoverage

import (
	"reflect"
	"testing"
)

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
