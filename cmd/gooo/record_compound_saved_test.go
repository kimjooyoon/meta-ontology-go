package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
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
	for _, path := range []string{"examples/local-field-bridge/trace.gooo", "examples/native-field-lineage/lineage.gooo"} {
		data, err := os.ReadFile("../../" + path)
		if err != nil || !bytes.Equal(data, entries["sources/"+path]) {
			t.Fatal("source identity", path, err)
		}
	}
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
