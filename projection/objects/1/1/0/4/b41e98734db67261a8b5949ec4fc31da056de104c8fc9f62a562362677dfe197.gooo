package main

import (
	"bytes"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"os"
	"path/filepath"
	"testing"
)

func TestRunProvenancePublishMalformedOrMissingInputDoesNotCreateLedger(t *testing.T) {
	fixture := newProvenanceCLIFixture(t)
	if err := os.WriteFile(fixture.evidencePath, []byte(`{"records":[]}`), 0o640); err != nil {
		t.Fatal(err)
	}
	output, code, stderr := fixture.publishRaw(t)
	if code != exitFailure || stderr != "" {
		t.Fatalf("empty evidence = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	response := decodeProvenanceResponse(t, output)
	if response.Status == provenanceStatusCommitted || response.Error == nil || response.Error.Code != "evidence.malformed" {
		t.Fatalf("empty evidence response = %#v", response)
	}
	if _, err := os.Stat(fixture.storePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("empty evidence changed store state: %v", err)
	}

	missing := filepath.Join(fixture.directory, "missing.gooo")
	args := []string{"provenance", "publish", "--json", missing, "--store", filepath.Join(fixture.directory, "missing.jsonl"), "--evidence", fixture.evidencePath}
	var stdout, missingStderr bytes.Buffer
	if code := run(args, &stdout, &missingStderr); code != exitFailure || missingStderr.Len() != 0 {
		t.Fatalf("missing source = code %d, stderr=%q, output=%s", code, missingStderr.String(), stdout.String())
	}
	missingResponse := decodeProvenanceResponse(t, stdout.Bytes())
	if missingResponse.Status == provenanceStatusCommitted || missingResponse.Error == nil || missingResponse.Error.Code != "source.read" {
		t.Fatalf("missing source response = %#v", missingResponse)
	}
	if _, err := os.Stat(filepath.Join(fixture.directory, "missing.jsonl")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing source changed store state: %v", err)
	}
}

type provenanceCLIFixture struct {
	directory      string
	sourcePath     string
	evidencePath   string
	storePath      string
	sourceDigest   string
	semanticDigest string
	graphDigest    string
	records        []provenance.Evidence
}
