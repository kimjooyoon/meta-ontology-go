package artifactemit

import (
	"strings"
	"testing"
)

func TestCompileSymbolicValueReaderProjection(t *testing.T) {
	result, err := CompileSymbolicValueReaderProjection(
		symbolicReaderFixture(), strings.Repeat("a", 40),
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != "PASS" || result.Resolution != "READER_PROJECTION_ONLY" {
		t.Fatalf("decision=%s resolution=%s", result.Decision, result.Resolution)
	}
	if result.Coordinates.Satisfied != 18 || result.Coordinates.Total != 18 {
		t.Fatalf("coordinates=%+v", result.Coordinates)
	}
	wantTotals := []int{5, 9, 11}
	if len(result.Readers) != len(wantTotals) {
		t.Fatalf("readers=%d", len(result.Readers))
	}
	for index, total := range wantTotals {
		if result.Readers[index].Coordinates.Total != total {
			t.Fatalf("reader[%d].total=%d", index, result.Readers[index].Coordinates.Total)
		}
	}
	if !symbolicReaderValidDigest(result.Digest) {
		t.Fatalf("digest=%q", result.Digest)
	}
}
