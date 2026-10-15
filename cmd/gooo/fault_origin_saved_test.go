package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestFaultOriginSavedSourceAndMembership(t *testing.T) {
	const directory = "../../docs/observations/fault-origins/"
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
		t.Fatal("reader archive identity", err)
	}
	entries := readNamedInputArchive(t, archive)
	if len(entries) != len(manifest.Members) {
		t.Fatal("reader archive membership")
	}
	for _, member := range manifest.Members {
		if len(entries[member.Path]) != member.Bytes || namedInputObservationSHA(entries[member.Path]) != member.SHA256 {
			t.Fatal("reader archived member", member.Path)
		}
	}
	const path = "examples/fault-origins/describe.gooo"
	source, err := os.ReadFile("../../" + path)
	if err != nil || !bytes.Equal(source, entries["sources/"+path]) {
		t.Fatal("reader source identity", err)
	}
}
