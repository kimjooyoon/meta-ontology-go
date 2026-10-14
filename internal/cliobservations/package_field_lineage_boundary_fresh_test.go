package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type fieldLineageAnswer14203 struct {
	Status     string `json:"status"`
	Roots      string `json:"roots"`
	RootCount  int    `json:"root_count"`
	Reachable  int    `json:"reachable_writes"`
	Parents    int    `json:"parent_links"`
	Reads      int    `json:"field_read_edges"`
	Unresolved int    `json:"unresolved_links"`
	Shared     int    `json:"shared_rhs_writes"`
	Scalar     int    `json:"scalar_rhs_writes"`
}

func syntheticFieldLineage14203() bodyexecution.FieldLineageColumns {
	return bodyexecution.FieldLineageColumns{
		SourceSHA256: "synthetic-field-equations14203", JournalSHA256: "synthetic-no-producer",
		SelectedUse: 2, RootAssignment: 3, RootClock: 4, WriteCount: 4, UseCount: 3, RecordingComplete: true,
		Bits: []int64{1, 2, 4, 8}, CompletionOrder: []int{0, 1, 2, 3}, Completions: []int{0, 1, 2, 3},
		Parents: []int{-1, -1, -1, 2}, UseStarts: []int{0, 0, 0, 2}, UseEnds: []int{0, 0, 2, 2},
		WriteCompleted: []bool{true, true, true, true}, WriteDrops: []int{0, 0, 0, 0},
		Modes:      []string{"parameter", "parameter", "scalar_rhs", "record_copy"},
		Kinds:      []string{"parameter", "parameter", "assignment", "assignment"},
		InputKinds: []string{"root_input", "root_input", "", ""}, Activities: []string{"merge", "merge", "merge", "merge"},
		Ports: []string{"input", "input", "", ""}, Records: []string{"record", "record", "record", "record"},
		Fields: []string{"left", "right", "left", "left"}, UseAssignments: []int{0, 1, 3}, UseClocks: []int{2, 2, 4},
	}
}

func TestGoooFieldLineageBoundaries14203(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	type trial struct {
		name   string
		change func(*bodyexecution.FieldLineageColumns)
		want   fieldLineageAnswer14203
	}
	complete := fieldLineageAnswer14203{"SCOPED_FIELD_ORIGINS", "merge|input|left;merge|input|right;", 2, 4, 1, 2, 0, 0, 1}
	invalid := fieldLineageAnswer14203{Status: "INVALID_INPUT"}
	trials := []trial{
		{"two-root-scalar-expression", nil, complete},
		{"shared-window-is-not-exact", func(c *bodyexecution.FieldLineageColumns) { c.Modes[2] = "shared_rhs" }, fieldLineageAnswer14203{Status: "PARTIAL_LINKS", Reachable: 2, Parents: 1, Unresolved: 1, Shared: 1}},
		{"known-recording-loss", func(c *bodyexecution.FieldLineageColumns) { c.RecordingComplete = false }, func() fieldLineageAnswer14203 { v := complete; v.Status = "PARTIAL_RECORDING"; return v }()},
		{"unresolved-field-use", func(c *bodyexecution.FieldLineageColumns) { c.UseAssignments[1] = -2 }, fieldLineageAnswer14203{Status: "PARTIAL_LINKS", Roots: "merge|input|left;", RootCount: 1, Reachable: 3, Parents: 1, Reads: 1, Unresolved: 1, Scalar: 1}},
		{"array-length-mismatch", func(c *bodyexecution.FieldLineageColumns) { c.Bits = c.Bits[:3] }, invalid},
		{"duplicate-completion-order", func(c *bodyexecution.FieldLineageColumns) { c.CompletionOrder[2] = 1 }, invalid},
		{"parent-crosses-field-identity", func(c *bodyexecution.FieldLineageColumns) { c.Fields[3] = "right" }, invalid},
		{"parent-is-from-future", func(c *bodyexecution.FieldLineageColumns) { c.Parents[2] = 3 }, invalid},
		{"use-before-completion", func(c *bodyexecution.FieldLineageColumns) { c.UseClocks[2] = 3; c.RootClock = 3 }, invalid},
		{"unresolved-selected-read", func(c *bodyexecution.FieldLineageColumns) { c.UseAssignments[2] = -1; c.RootAssignment = -1 }, fieldLineageAnswer14203{Status: "UNRESOLVED_READ"}},
	}
	var rows []map[string]bodyexecution.FieldLineageColumns
	for _, tc := range trials {
		c := syntheticFieldLineage14203()
		if tc.change != nil {
			tc.change(&c)
		}
		rows = append(rows, map[string]bodyexecution.FieldLineageColumns{"field/lineage:Trace": c})
	}
	input, err := json.Marshal(map[string]any{"schema": bodyexecution.CompositionInputsSchema, "inputs": rows})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "synthetic14203.json")
	if err := os.WriteFile(path, input, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--inputs", path, filepath.Join(root, "examples/native-field-lineage/gooo.workspace.json")}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err, receipt.Error)
	}
	runtime := receipt.Result.Runtime
	if receipt.Decision != "OBSERVED" || runtime.ModelCalls != 0 || runtime.FiniteTotal != 0 || len(runtime.Traces) != len(trials) {
		t.Fatal("input-only synthetic scope", receipt.Decision)
	}
	for i, trace := range runtime.Traces {
		var got fieldLineageAnswer14203
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &got) != nil || got != trials[i].want {
			t.Fatal(trials[i].name, got, trials[i].want)
		}
	}
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		var compressed bytes.Buffer
		z := gzip.NewWriter(&compressed)
		if _, err := z.Write(out.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "synthetic14203.json.gz"), compressed.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("10 fresh synthetic journals->Gooo joins2 roots, exposes shared/unknown/dropped scope, rejects5 structural contradictions;original producer executions0;models0;input-only finite0")
}
