package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckSemanticProvenanceRejectsStaleSourceWithoutAppend(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "main.gooo")
	storePath := filepath.Join(directory, "ledger.jsonl")
	if err := os.WriteFile(sourcePath, []byte(validSource), 0o640); err != nil {
		t.Fatal(err)
	}
	if output, code, stderr := runCheckProvenanceJSON(t, sourcePath, storePath); code != exitOK || stderr != "" {
		t.Fatalf("setup check = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	before := mustReadProvenanceFile(t, storePath)
	if err := os.WriteFile(sourcePath, append([]byte(validSource), '\n'), 0o640); err != nil {
		t.Fatal(err)
	}
	output, code, stderr := runCheckProvenanceJSON(t, sourcePath, storePath)
	if code != exitFailure || stderr != "" {
		t.Fatalf("stale check = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	report := decodeCheckJSON(t, output)
	if report.Provenance == nil || report.Provenance.CheckStatus != checkStatusPass || report.Provenance.Status == provenanceStatusCommitted || report.Provenance.Error == nil || report.Provenance.Error.Code != "provenance.stale-source" {
		t.Fatalf("stale check report = %#v", report)
	}
	if after := mustReadProvenanceFile(t, storePath); !bytes.Equal(after, before) {
		t.Fatal("stale source appended to the existing ledger")
	}
}
