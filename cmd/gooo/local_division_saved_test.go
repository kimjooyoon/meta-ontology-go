package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestLocalDivisionSavedProvenance(t *testing.T) {
	const directory = "../../docs/observations/local-division-origins/"
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
		t.Fatal("archive population")
	}
	for _, m := range manifest.Members {
		if len(entries[m.Path]) != m.Bytes || namedInputObservationSHA(entries[m.Path]) != m.SHA256 {
			t.Fatal("archive member", m.Path)
		}
	}
	source, err := os.ReadFile("../../examples/compound-updates/main.gooo")
	if err != nil || !bytes.Equal(source, entries["sources/examples/compound-updates/main.gooo"]) {
		t.Fatal("public source binding", err)
	}
	for _, c := range []struct {
		path           string
		normal, faults int
	}{
		{"repair-outputs/local-division-faults.json", 3, 2},
		{"mixed-outputs/mixed-division-faults.json", 2, 2},
		{"public-outputs/public-local-division.json", 2, 1},
	} {
		var saved struct {
			Source  string
			Runtime bodyexecution.CompositionRuntime
		}
		if err := json.Unmarshal(entries[c.path], &saved); err != nil {
			t.Fatal(err)
		}
		r := saved.Runtime
		if r.ModelCalls != 0 || r.FinitePassed != c.normal || r.FiniteTotal != c.normal+c.faults || r.Outcomes == nil || r.Outcomes.Faulted != c.faults || r.SelectedSourceSHA256 != "sha256:"+namedInputObservationSHA([]byte(saved.Source)) {
			t.Fatal("saved scope or source changed", c.path)
		}
	}
}
