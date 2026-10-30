package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplicitRecordLibraryRetainsSourceAndPriorOutput(t *testing.T) {
	root := t.TempDir()
	source, err := os.ReadFile("../../internal/meta/runtimeidentity/fields.gooo")
	if err != nil {
		t.Fatal(err)
	}
	target, err := selectedTarget("fields.gooo", "fields.go", "urn:gooo:runtime-identity:describe", "runtimeidentity")
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(root, target.source), source)
	if err := projectLibrary(root, target, true); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(root, target.output))
	changed := bytes.Replace(source, []byte("CASE_CONTENT_DIGEST"), []byte("ChangedSourceRole"), 1)
	writeFixture(t, filepath.Join(root, target.source), changed)
	if err := projectLibrary(root, target, false); err == nil {
		t.Fatal("changed source accepted without regeneration")
	}
	after, _ := os.ReadFile(filepath.Join(root, target.output))
	if !bytes.Equal(before, after) {
		t.Fatal("check mode changed the prior projection")
	}
	if err := projectLibrary(root, target, true); err != nil {
		t.Fatal(err)
	}
	regenerated, _ := os.ReadFile(filepath.Join(root, target.output))
	if !bytes.Contains(regenerated, []byte("ChangedSourceRole")) || bytes.Equal(before, regenerated) {
		t.Fatal("record-valued source change was not projected")
	}
}

func TestProjectionRefusesAPartialTargetAndItsOwnInputAsOutput(t *testing.T) {
	if _, err := selectedTarget("fields.gooo", "", "", ""); err == nil {
		t.Fatal("partial target could overwrite the default version projection")
	}
	root := t.TempDir()
	path := filepath.Join(root, "source.gooo")
	before := []byte("retain original source\n")
	writeFixture(t, path, before)
	target := projectionTarget{"source.gooo", "./source.gooo", "urn:example:entry", "example"}
	if err := projectLibrary(root, target, true); err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("source/output alias accepted: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("refused projection changed the original source", err)
	}
}
