package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestRecordedBridgeContract15503(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	load := func(path string) bodyexecution.CompositionRuntime {
		t.Helper()
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		z, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		defer z.Close()
		raw, err := io.ReadAll(z)
		if err != nil {
			t.Fatal(err)
		}
		var r bodyexecution.CompositionRuntime
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	fixture := filepath.Join(root, "examples/local-field-bridge/recorded-runtime.json.gz")
	base := load(fixture)
	query := bodyexecution.NativeOriginQuery{Kind: "field-uses", Limit: 1}
	columns, report, err := bodyexecution.SelectRecordedBridgeLineage(base, query)
	if err != nil || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 || report.BridgeObservation == nil {
		t.Fatal(err, report)
	}
	encoded, err := json.Marshal(columns)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil || len(fields) != 32 {
		t.Fatal("native primitive record32 profile", err, len(fields))
	}
	scalars, literals := 0, 0
	for i, kind := range columns.Kinds {
		if columns.Modes[i] == "literal" {
			literals++
		}
		if kind == "scalar_local" {
			if columns.Records[i] != "" || columns.Fields[i] != "" || columns.NodeNames[i] == "" || columns.SourceOrdinals[i] < 0 || columns.TypeIDs[i] == "" {
				t.Fatal("scalar source identity fabricated as a record field")
			}
			scalars++
		}
	}
	if scalars != len(base.Traces[0].LocalWrites) || literals == 0 {
		t.Fatal("source scalar and literal observations", scalars, literals)
	}
	mutations := []struct {
		name   string
		change func(*bodyexecution.CompositionRuntime)
	}{
		{"duplicate-shared-completion", func(r *bodyexecution.CompositionRuntime) {
			r.Traces[0].Bridge.Locals[0].Completion = r.Traces[0].Bridge.Fields[0].Completion
		}},
		{"field-order-contradicts-own-clock", func(r *bodyexecution.CompositionRuntime) {
			b := r.Traces[0].Bridge
			b.Fields[0].Completion, b.Fields[1].Completion = b.Fields[1].Completion, b.Fields[0].Completion
		}},
		{"local-read-before-its-write", func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Bridge.LocalClocks[0] = 0 }},
		{"future-cross-window", func(r *bodyexecution.CompositionRuntime) {
			b := &r.Traces[0].Bridge.Locals[0]
			b.OtherStart, b.OtherEnd = len(r.Traces[0].FieldUses)-1, len(r.Traces[0].FieldUses)
		}},
		{"start-after-completion", func(r *bodyexecution.CompositionRuntime) {
			r.Traces[0].Bridge.Locals[0].StartClock = r.Traces[0].Bridge.Completions
		}},
		{"read-clock-population", func(r *bodyexecution.CompositionRuntime) {
			r.Traces[0].Bridge.FieldClocks = r.Traces[0].Bridge.FieldClocks[:1]
		}},
		{"summary-population", func(r *bodyexecution.CompositionRuntime) { r.BridgeObservation.CrossReads++ }},
		{"bridge-summary-missing", func(r *bodyexecution.CompositionRuntime) { r.BridgeObservation = nil }},
		{"source-reference-missing", func(r *bodyexecution.CompositionRuntime) { r.BridgeObservation.SHA256 = "" }},
	}
	for _, tc := range mutations {
		r := load(fixture)
		tc.change(&r)
		if _, _, err := bodyexecution.SelectRecordedBridgeLineage(r, query); err == nil {
			t.Fatal("recorded bridge contradiction accepted", tc.name)
		}
	}
	old := load(filepath.Join(root, "examples/native-field-lineage/recorded-runtime.json.gz"))
	if _, _, err := bodyexecution.SelectRecordedBridgeLineage(old, query); err == nil {
		t.Fatal("bridge invented for old field-only record")
	}
	if _, _, err := bodyexecution.SelectRecordedNativeOrigins(old, query); err != nil {
		t.Fatal("old field-only facts unavailable", err)
	}
	t.Log("saved native bridge maps32 columns with true scalar identities and literal facts;nine fresh structural contradictions rejected;old field-only rows remain readable without invented bridge;native/model executions0")
}
