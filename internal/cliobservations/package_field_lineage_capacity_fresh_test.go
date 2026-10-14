package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func fieldLineageChain14611(writes, uses int, reverse bool) bodyexecution.FieldLineageColumns {
	c := bodyexecution.FieldLineageColumns{SourceSHA256: "synthetic-capacity14611", JournalSHA256: "synthetic-chain-no-producer", RootAssignment: writes - 1, RootClock: writes, SelectedUse: uses - 1, WriteCount: writes, UseCount: uses, RecordingComplete: true}
	for i := range writes {
		parent, completion, ordered, kind, inputKind := i-1, i, i, "assignment", ""
		if reverse {
			parent, completion, ordered = i+1, writes-1-i, writes-1-i
		}
		if completion == 0 {
			parent, kind, inputKind = -1, "parameter", "root_input"
		}
		c.Bits = append(c.Bits, int64(1)<<i)
		c.CompletionOrder, c.Completions, c.Parents = append(c.CompletionOrder, ordered), append(c.Completions, completion), append(c.Parents, parent)
		c.UseStarts, c.UseEnds = append(c.UseStarts, 0), append(c.UseEnds, 0)
		c.WriteCompleted, c.WriteDrops = append(c.WriteCompleted, true), append(c.WriteDrops, 0)
		c.Modes, c.Kinds, c.InputKinds = append(c.Modes, "record_copy"), append(c.Kinds, kind), append(c.InputKinds, inputKind)
		c.Activities, c.Ports, c.Records, c.Fields = append(c.Activities, "edge"), append(c.Ports, "input"), append(c.Records, "record"), append(c.Fields, "one")
	}
	if reverse {
		c.RootAssignment = 0
	}
	for range uses {
		c.UseAssignments, c.UseClocks = append(c.UseAssignments, c.RootAssignment), append(c.UseClocks, c.RootClock)
	}
	return c
}

func TestGoooFieldLineageCapacity14611(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	full := fieldLineageAnswer14203{Status: "SCOPED_FIELD_ORIGINS", Roots: "edge|input|one;", RootCount: 1, Reachable: 60, Parents: 59}
	invalid := fieldLineageAnswer14203{Status: "INVALID_INPUT"}
	wrongBit := fieldLineageChain14611(60, 128, false)
	wrongBit.Bits[59] = wrongBit.Bits[58]
	interrupted := fieldLineageChain14611(60, 128, false)
	interrupted.RootAssignment, interrupted.RootClock, interrupted.RecordingComplete = 58, 59, false
	interrupted.WriteCompleted[59], interrupted.Completions[59], interrupted.CompletionOrder[59] = false, -1, -1
	for i := range interrupted.UseAssignments {
		interrupted.UseAssignments[i], interrupted.UseClocks[i] = 58, 59
	}
	cases := []struct {
		name string
		c    bodyexecution.FieldLineageColumns
		want fieldLineageAnswer14203
	}{
		{"60writes128uses", fieldLineageChain14611(60, 128, false), full},
		{"reservation-order-opposes-completion", fieldLineageChain14611(60, 128, true), full},
		{"61writes", fieldLineageChain14611(61, 128, false), invalid},
		{"129uses", fieldLineageChain14611(60, 129, false), invalid},
		{"duplicate-high-bit", wrongBit, invalid},
		{"unfinished-final-write", interrupted, fieldLineageAnswer14203{Status: "PARTIAL_RECORDING", Roots: "edge|input|one;", RootCount: 1, Reachable: 59, Parents: 58}},
	}
	var rows []map[string]bodyexecution.FieldLineageColumns
	for _, c := range cases {
		rows = append(rows, map[string]bodyexecution.FieldLineageColumns{"field/lineage:Trace": c.c})
	}
	for start := 0; start < len(cases); start += 2 {
		end := start + 2
		raw, err := json.Marshal(map[string]any{"schema": bodyexecution.CompositionInputsSchema, "inputs": rows[start:end]})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "capacity14611.json")
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		var out, diagnostics bytes.Buffer
		if code := run([]string{"package", "execute", "--json", "--inputs", path, filepath.Join(root, "examples/native-field-lineage/gooo.workspace.json")}, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String(), out.String())
		}
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Decision != "OBSERVED" || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 || len(receipt.Result.Runtime.Traces) != end-start {
			t.Fatal("capacity observation scope", err, receipt.Error)
		}
		for i, trace := range receipt.Result.Runtime.Traces {
			var got fieldLineageAnswer14203
			if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &got) != nil || got != cases[start+i].want {
				t.Fatal(cases[start+i].name, got, cases[start+i].want)
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
			if err := os.WriteFile(filepath.Join(dir, "capacity14611-batch"+strconv.Itoa(start/2)+".json.gz"), compressed.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("6 fresh synthetic boundary journals;60writes128uses/high-bit traversal and reversed completion order;61write/129use rejection;unfinished write stays partial;producer executions0;models0;input-only finite0")
}
