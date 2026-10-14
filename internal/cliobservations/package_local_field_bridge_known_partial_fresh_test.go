package cliobservations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestGoooKnownRootsWithinPartial15817(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	var rows []map[string]bodyexecution.BridgeLineageColumns
	for i := range 3 {
		c := knownPartialColumns15817()
		if i == 1 {
			c.Modes[1] = "scalar_rhs"
		}
		if i == 2 {
			c.RecordingComplete = false
		}
		rows = append(rows, map[string]bodyexecution.BridgeLineageColumns{"origin/bridgetrace:Trace": c})
	}
	raw, err := json.Marshal(map[string]any{"schema": bodyexecution.CompositionInputsSchema, "inputs": rows})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "known-partial15817.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	answers := runBridgeConsumer15717(t, "known-partial15817", []string{"package", "execute", "--json", "--inputs", path, filepath.Join(root, "examples/local-field-bridge/trace.workspace.json")})
	if len(answers) != 3 {
		t.Fatal("known-partial case count", len(answers))
	}
	for i, got := range answers {
		want := bridgeAnswer15717{Collection: 1, Status: "PARTIAL_LINKS", Roots: "mixed|input|input-value;", RootCount: 1, Carrier: 1, Unresolved: 1, Complete: true}
		if i == 1 {
			want.Status, want.Unresolved, want.Collection = "SCOPED_FIELD_LOCAL_ORIGINS", 0, 0
		}
		if i == 2 {
			want.Status, want.Complete = "PARTIAL_RECORDING", false
		}
		if got != want {
			t.Fatal("known roots were discarded when another source is unknown", i, got, want)
		}
	}
	t.Log("3 fresh synthetic mixed-window cases->Gooo3/3;known field/local origin retained beside untracked collection input and beside recording loss;normal known-only case remains scoped;producer/model executions0;input-only finite0")
}

func knownPartialColumns15817() bodyexecution.BridgeLineageColumns {
	return bodyexecution.BridgeLineageColumns{
		SourceSHA256: "synthetic-mixed-window15817", JournalSHA256: "synthetic-no-native-producer", SelectedUse: 1, RootAssignment: 2, RootClock: 3, WriteCount: 3, UseCount: 3, RecordingComplete: true,
		Bits: []int64{1, 2, 4}, CompletionOrder: []int{0, 1, 2}, Completions: []int{0, 1, 2}, Parents: []int{-1, -1, -1}, UseStarts: []int{0, 2, 1}, UseEnds: []int{0, 2, 1},
		WriteCompleted: []bool{true, true, true}, WriteDrops: []int{0, 0, 0}, Modes: []string{"call_parameter_or_input", "collection_rhs", "scalar_rhs"}, Kinds: []string{"parameter", "scalar_local", "field_assignment"},
		InputKinds: []string{"root_input", "", ""}, Activities: []string{"mixed", "mixed", "mixed"}, Ports: []string{"input", "", ""}, Records: []string{"input-record", "", "output-record"}, Fields: []string{"input-value", "", "value"}, UseAssignments: []int{0, 2, 1}, UseClocks: []int{1, 3, 2},
		OtherStarts: []int{2, 0, 2}, OtherEnds: []int{2, 1, 3}, NodeNames: []string{"input", "cached", "output"}, SourceOrdinals: []int{-1, 0, 1}, TypeIDs: []string{"urn:gooo:type:integer", "urn:gooo:type:integer", "urn:gooo:type:integer"}, FieldUseCount: 2}
}
