package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type bridgeAnswer15717 struct {
	Collection int    `json:"collection_rhs_writes"`
	Status     string `json:"status"`
	Roots      string `json:"roots"`
	RootCount  int    `json:"root_count"`
	Carrier    int    `json:"carrier_writes"`
	Literals   int    `json:"literal_writes"`
	Unresolved int    `json:"unresolved_links"`
	Complete   bool   `json:"recording_complete"`
}

func TestGoooBridgePartialAndLeaf15717(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	file, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/bridge-fault15617-runtime.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
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
	saved := filepath.Join(t.TempDir(), "fault15717.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(root, "examples/local-field-bridge/trace.workspace.json")
	offset := 0
	for c, trace := range r.Traces {
		selected := 0
		if c == 0 {
			selected = len(trace.FieldUses) - 1
		}
		answers := runBridgeConsumer15717(t, "fault15717-case"+strconv.Itoa(c), []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-shape", "field-local-columns", "--origin-offset", strconv.Itoa(offset + selected), manifest})
		want := bridgeAnswer15717{Status: "PARTIAL_RECORDING", Roots: "urn:bridgefault15617:build|input|urn:bridgefault15617:value;", RootCount: 1}
		if c == 0 {
			want = bridgeAnswer15717{Status: "SCOPED_FIELD_LOCAL_ORIGINS", Roots: "urn:bridgefault15617:build|input|urn:bridgefault15617:value;urn:bridgefault15617:build|input|urn:bridgefault15617:divisor;", RootCount: 2, Complete: true}
		}
		if len(answers) != 1 || answers[0] != want {
			t.Fatal("Gooo per-case native interruption scope", c, answers, want)
		}
		offset += len(trace.FieldUses)
	}
	var rows []map[string]bodyexecution.BridgeLineageColumns
	var expected []bridgeAnswer15717
	for i := range 6 {
		c := syntheticBridgeLeaf15717()
		want := bridgeAnswer15717{Status: "SCOPED_FIELD_LOCAL_ORIGINS", Carrier: 1, Literals: 1, Complete: true}
		switch i {
		case 1:
			c.Modes[0] = "unresolved_rhs"
			want.Status, want.Literals, want.Unresolved = "PARTIAL_LINKS", 0, 1
		case 2:
			c.Modes[0] = "collection_rhs"
			want.Collection = 1
			want.Status, want.Literals, want.Unresolved = "PARTIAL_LINKS", 0, 1
		case 3:
			c.RecordingComplete = false
			want.Status, want.Complete = "PARTIAL_RECORDING", false
		case 4:
			c.Records[0] = "invented-scalar-record"
			want = bridgeAnswer15717{Status: "INVALID_INPUT", Complete: true}
		case 5:
			c.UseClocks[1] = 0
			want = bridgeAnswer15717{Status: "INVALID_INPUT", Complete: true}
		}
		rows = append(rows, map[string]bodyexecution.BridgeLineageColumns{"origin/bridgetrace:Trace": c})
		expected = append(expected, want)
	}
	input, err := json.Marshal(map[string]any{"schema": bodyexecution.CompositionInputsSchema, "inputs": rows})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "synthetic15717.json")
	if err := os.WriteFile(path, input, 0600); err != nil {
		t.Fatal(err)
	}
	answers := runBridgeConsumer15717(t, "synthetic15717", []string{"package", "execute", "--json", "--inputs", path, manifest})
	if len(answers) != len(expected) {
		t.Fatal("synthetic result count", len(answers))
	}
	for i, got := range answers {
		if got != expected[i] {
			t.Fatal("synthetic leaf scope", i, got, expected[i])
		}
	}
	t.Log("3 new saved native fault-context queries->Gooo case-scoped complete/partial3/3;6 new synthetic literal/unknown/collection/partial/invalid inputs->Gooo6/6;original producer executions0;models0;input-only finite0")
}

func syntheticBridgeLeaf15717() bodyexecution.BridgeLineageColumns {
	return bodyexecution.BridgeLineageColumns{
		SourceSHA256: "synthetic-no-native-producer15717", JournalSHA256: "synthetic-leaf-only", RootAssignment: 1, RootClock: 2, WriteCount: 2, UseCount: 2, RecordingComplete: true,
		Bits: []int64{1, 2}, CompletionOrder: []int{0, 1}, Completions: []int{0, 1}, Parents: []int{-1, -1}, UseStarts: []int{1, 0}, UseEnds: []int{1, 0},
		WriteCompleted: []bool{true, true}, WriteDrops: []int{0, 0}, Modes: []string{"literal", "scalar_rhs"}, Kinds: []string{"scalar_local", "field_assignment"},
		InputKinds: []string{"", ""}, Activities: []string{"synthetic", "synthetic"}, Ports: []string{"", ""}, Records: []string{"", "record"}, Fields: []string{"", "value"}, UseAssignments: []int{1, 0}, UseClocks: []int{2, 1},
		OtherStarts: []int{0, 1}, OtherEnds: []int{0, 2}, NodeNames: []string{"seed", "result"}, SourceOrdinals: []int{0, 1}, TypeIDs: []string{"urn:gooo:type:integer", "urn:gooo:type:integer"}, FieldUseCount: 1}
}

func runBridgeConsumer15717(t *testing.T, name string, args []string) []bridgeAnswer15717 {
	t.Helper()
	var out, diagnostic bytes.Buffer
	if code := run(args, &out, &diagnostic); code != exitOK {
		t.Fatal(name, code, diagnostic.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Decision != "OBSERVED" || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("Gooo observed result scope", name, err, receipt.Error)
	}
	if receipt.OriginInput != nil && (receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0) {
		t.Fatal("producer was executed while reading", name)
	}
	var answers []bridgeAnswer15717
	for _, trace := range receipt.Result.Runtime.Traces {
		var a bridgeAnswer15717
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &a) != nil {
			t.Fatal("Gooo answer shape", name)
		}
		answers = append(answers, a)
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
		if err := os.WriteFile(filepath.Join(dir, name+".json.gz"), compressed.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return answers
}
