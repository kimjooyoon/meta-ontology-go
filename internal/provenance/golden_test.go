package provenance

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBillingGoldenJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "golden.jsonl")
	if err := New(path).Append(BillingFixture()...); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/golden.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, golden) {
		t.Fatalf("billing ledger changed from deterministic golden:\n got %s\nwant %s", data, golden)
	}
}
