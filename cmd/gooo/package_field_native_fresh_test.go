package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestWorkspaceNativeFields13517(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	root := filepath.Join(sourceRoot, "examples/native-record-fields")
	call := func(name string, args ...string) []byte {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(name, code, diagnostics.String(), out.String())
		}
		if dir := os.Getenv("GOOO_READ_OUTPUT"); dir != "" {
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
		return append([]byte(nil), out.Bytes()...)
	}
	producer := call("field-producer13517", "package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json"))
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(producer, &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err, receipt.Error)
	}
	runtime := receipt.Result.Runtime
	if runtime.FinitePassed != 4 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 || runtime.FieldObservation == nil || runtime.FieldObservation.Status != "COMPLETE" || runtime.FieldObservation.ParentLinks == 0 || runtime.FieldObservation.UnknownParents != 0 {
		t.Fatal("actual workspace field observation", runtime.FinitePassed, runtime.FieldObservation)
	}
	rows, report, err := bodyexecution.SelectRecordedNativeOrigins(runtime, bodyexecution.NativeOriginQuery{Kind: "field-uses", Limit: 8})
	if err != nil || len(rows) != 8 || report.FieldObservation == nil || !report.Truncated || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("saved field page", err, report)
	}
	for _, raw := range rows {
		var fact bodyexecution.CompositionFieldFact
		if err := json.Unmarshal(raw, &fact); err != nil || !fact.HasAssignment || !fact.Completed || fact.RecordID == "" || fact.FieldID == "" || fact.ProjectionSHA256 == "" {
			t.Fatal("flat actual field reference", err, fact)
		}
	}
	path := filepath.Join(t.TempDir(), "field13517-receipt.json")
	if err := os.WriteFile(path, producer, 0600); err != nil {
		t.Fatal(err)
	}
	consumer := call("field-consumer13517", "package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "field-uses", "--origin-limit", "8", filepath.Join(root, "reader.workspace.json"))
	var observed packageExecutionReceipt
	if err := json.Unmarshal(consumer, &observed); err != nil || observed.Result == nil || observed.Decision != "OBSERVED" {
		t.Fatal(err, observed.Error)
	}
	if observed.Result.Runtime.FiniteTotal != 0 || observed.Result.Runtime.ModelCalls != 0 || len(observed.Result.Runtime.Traces) != 8 || observed.OriginInput == nil || observed.OriginInput.ProducerExecutions != 0 {
		t.Fatal("Gooo field reader scope")
	}
	want := []string{
		"input:urn:gooo:native-fields:input-left", "input:urn:gooo:native-fields:input-right",
		"observed:urn:gooo:native-fields:left", "observed:urn:gooo:native-fields:right",
		"linked:urn:gooo:native-fields:left", "linked:urn:gooo:native-fields:right",
		"linked:urn:gooo:native-fields:left", "observed:urn:gooo:native-fields:left",
	}
	for i, trace := range observed.Result.Runtime.Traces {
		var actual string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual != want[i] {
			t.Fatal("Gooo field description", i, actual, want[i])
		}
	}
	t.Log("actual new workspace4/4,field-sidecar2-run equality,complete journal/known parents;8 saved facts->real32-fieldGooo reader8/8 external comparisons;producer executions during reads0,models0")
}
