package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckSemanticProvenanceRejectsMalformedNegativeAndMissingParentWithoutWrite(t *testing.T) {
	directory := t.TempDir()
	malformedPath := filepath.Join(directory, "malformed.gooo")
	negativePath := filepath.Join(directory, "negative.gooo")
	if err := os.WriteFile(malformedPath, []byte("package billing\nnamespace billing\nentity Broken id \"x\" @\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(negativePath, []byte("package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Missing) -> Order\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	for _, filename := range []string{malformedPath, negativePath} {
		storePath := filepath.Join(directory, filepath.Base(filename)+".jsonl")
		var stdout, stderr bytes.Buffer
		if code := run([]string{"check", "--semantic", "--provenance-store", storePath, filename}, &stdout, &stderr); code != exitFailure {
			t.Fatalf("negative check %q code = %d, stdout=%q, stderr=%q", filename, code, stdout.String(), stderr.String())
		}
		if _, err := os.Stat(storePath); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("negative check %q wrote store: %v", filename, err)
		}
	}
	missingParent := filepath.Join(directory, "missing", "ledger.jsonl")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check", "--semantic", "--provenance-store", missingParent, negativePath}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("missing-parent check code = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Dir(missingParent)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing-parent check created parent: %v", err)
	}
}
func TestRunCheckProvenanceFlagRequiresSemanticAndDefaultRemainsReadOnly(t *testing.T) {
	directory := t.TempDir()
	storePath := filepath.Join(directory, "ledger.jsonl")
	var stdout, stderr bytes.Buffer
	if code := runCheck([]string{"--provenance-store", storePath, "fixture.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr); code != exitUsage || stderr.String() != checkUsage+"\n" {
		t.Fatalf("invalid combination = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(storePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("invalid combination wrote store: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runCheck([]string{"--semantic", "fixture.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK || stdout.String() != "ok: fixture.gooo\n" || stderr.String() != deferredCheckProvenance+"\n" {
		t.Fatalf("default semantic check = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}
