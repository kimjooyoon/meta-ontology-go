package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestTypedAncestryBounds13201(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	makeColumns := func(n int) bodycodegen.TypedOriginColumns {
		c := bodycodegen.TypedOriginColumns{SourceSHA256: "synthetic-bound-input13201", Count: n, Root: uint16(n),
			IDs: make([]uint16, n), Bits: make([]int64, n), ParentA: make([]uint16, n), ParentB: make([]uint16, n),
			Guards: make([]uint16, n), Conditions: make([]uint16, n), Kinds: make([]string, n), FieldIDs: make([]string, n), InputPorts: make([]string, n)}
		for i := range n {
			c.IDs[i], c.Bits[i], c.Kinds[i] = uint16(i+1), int64(1)<<i, "expression"
			if i > 0 {
				c.ParentA[i] = uint16(i)
			}
		}
		if n > 0 {
			c.Kinds[0], c.FieldIDs[0], c.InputPorts[0] = "input", "urn:typed13201:field", "input"
		}
		return c
	}
	columns := []bodycodegen.TypedOriginColumns{makeColumns(1), makeColumns(60), makeColumns(0), makeColumns(61), makeColumns(1), makeColumns(1), makeColumns(1)}
	columns[4].Bits[0] = 3
	columns[5].IDs[0] = 2
	columns[6].ParentA[0] = 1
	rows := make([]map[string]bodycodegen.TypedOriginColumns, len(columns))
	for i, c := range columns {
		rows[i] = map[string]bodycodegen.TypedOriginColumns{"typed/ancestry:Inputs": c}
	}
	raw, _ := json.Marshal(struct {
		Schema string                                      `json:"schema"`
		Inputs []map[string]bodycodegen.TypedOriginColumns `json:"inputs"`
	}{"gooo/body-composition-inputs/v1", rows})
	path := filepath.Join(t.TempDir(), "bounds13201.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--inputs", path, filepath.Join(sourceRoot, "examples/typed-value-origins/ancestry.workspace.json")}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	if dir := os.Getenv("GOOO_TYPED_OUTPUT"); dir != "" {
		var compressed bytes.Buffer
		z := gzip.NewWriter(&compressed)
		if _, err := z.Write(out.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "consumer13201.json.gz"), compressed.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var observed packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &observed); err != nil || observed.Result == nil || observed.Decision != "OBSERVED" || len(observed.Result.Runtime.Traces) != 7 || observed.Result.Runtime.FiniteTotal != 0 || observed.Result.Runtime.ModelCalls != 0 {
		t.Fatal("bounded input-only scope", err)
	}
	for i, trace := range observed.Result.Runtime.Traces {
		want := "unusable-input"
		if i < 2 {
			want = "input:urn:typed13201:field;"
		}
		var actual string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual != want {
			t.Fatal("Gooo boundary outcome", i, actual, want)
		}
	}
	t.Log("7 fresh synthetic Gooo graph inputs:1/60 node ancestry,0/61 node refusal,bit/order/self-edge refusal;input-only finite0,models0")
}
