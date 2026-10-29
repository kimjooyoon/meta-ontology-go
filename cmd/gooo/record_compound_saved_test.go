package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestRecordCompoundSavedOrigins(t *testing.T) {
	const directory = "../../docs/observations/record-field-updates/"
	raw, err := os.ReadFile(directory + "archive-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest namedInputArchiveManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(directory + manifest.Archive)
	if err != nil || namedInputObservationSHA(archive) != manifest.SHA256 {
		t.Fatal("archive identity", err)
	}
	entries := readNamedInputArchive(t, archive)
	if len(entries) != len(manifest.Members) {
		t.Fatal("archive membership")
	}
	for _, member := range manifest.Members {
		data, ok := entries[member.Path]
		if !ok || len(data) != member.Bytes || namedInputObservationSHA(data) != member.SHA256 {
			t.Fatal("saved observation changed", member.Path)
		}
	}
	for _, path := range []string{"examples/native-field-lineage/lineage.gooo"} {
		data, err := os.ReadFile("../../" + path)
		if err != nil || !bytes.Equal(data, entries["sources/"+path]) {
			t.Fatal("source identity", path, err)
		}
	}
	checkArchivedCompoundBridgeCompatibility(t, entries["sources/examples/local-field-bridge/trace.gooo"])
	if !bytes.Equal(readCompoundRuntimeFixture(t), entries["recorded-runtime-repaired.json"]) {
		t.Fatal("example receipt differs from retained repair")
	}
	var runtime struct {
		SelectedSourceSHA256 string `json:"selected_source_sha256"`
	}
	if err := json.Unmarshal(readCompoundRuntimeFixture(t), &runtime); err != nil || runtime.SelectedSourceSHA256 != "sha256:"+namedInputObservationSHA(entries["sources/examples/compound-updates/main.gooo"]) {
		t.Fatal("historical record example source", err)
	}
}

// Historical observations retain their original source bytes. A current
// reader may evolve while preserving the old input and output contract.
func checkArchivedCompoundBridgeCompatibility(t *testing.T, historical []byte) {
	t.Helper()
	current, err := os.ReadFile("../../examples/local-field-bridge/trace.gooo")
	if err != nil || len(historical) == 0 {
		t.Fatal("historical or current bridge source missing", err)
	}
	var runtime bodyexecution.CompositionRuntime
	if err := json.Unmarshal(readCompoundRuntimeFixture(t), &runtime); err != nil {
		t.Fatal(err)
	}
	rows, report, err := bodyexecution.SelectRecordedBridgeLineagePage(runtime,
		bodyexecution.NativeOriginQuery{Kind: "field-uses", Offset: 9, Limit: 4})
	if err != nil || len(rows) != 4 || report.ProducerExecutions != 0 {
		t.Fatal("saved compound observation transport", err, report)
	}
	for i, row := range rows {
		input, err := json.Marshal([]any{row})
		if err != nil {
			t.Fatal(err)
		}
		before, err := bodycodegen.EvaluateFixedValue(context.Background(), "historical-trace.gooo", historical, "Trace", input)
		if err != nil {
			t.Fatal("historical reader", i, err)
		}
		after, err := bodycodegen.EvaluateFixedValue(context.Background(), "current-trace.gooo", current, "Trace", input)
		if err != nil || before.ActivityID != after.ActivityID || !bytes.Equal(before.Value, after.Value) {
			t.Fatal("legacy reader contract changed", i, err, string(before.Value), string(after.Value))
		}
	}
}
