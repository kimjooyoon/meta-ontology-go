package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckSemanticProvenanceSeparatesCheckPassFromRejectedCommit(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "main.gooo")
	storePath := filepath.Join(directory, "ledger.jsonl")
	if err := os.WriteFile(sourcePath, []byte(validSource), 0o640); err != nil {
		t.Fatal(err)
	}

	file, diagnostics := syntax.ParseFile(sourcePath, validSource)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	ir, err := semanticCheckIR(file, commandDeadline)
	if err != nil {
		t.Fatal(err)
	}
	conflict := semanticCheckEvidence(sourcePath, file, semantic.StableHash([]byte(validSource)), ir.StableHash(), ir.Graph.StableHash())
	conflict.Status = provenance.StatusCandidate
	conflict.Attributes = map[string]string{"check_schema": semanticCheckSchemaVersion, "result": "deferred"}
	if err := provenance.New(storePath).Append(conflict); err != nil {
		t.Fatalf("seed conflicting evidence = %v", err)
	}
	before := mustReadProvenanceFile(t, storePath)
	output, rejectedCode, rejectedStderr := runCheckProvenanceJSON(t, sourcePath, storePath)
	if rejectedCode != exitFailure || rejectedStderr != "" {
		t.Fatalf("rejected check = code %d, stderr=%q, output=%s", rejectedCode, rejectedStderr, output)
	}
	rejected := decodeCheckJSON(t, output)
	if rejected.Status != "error" || rejected.SemanticHash != ir.StableHash() || rejected.Provenance == nil || rejected.Provenance.CheckStatus != checkStatusPass || rejected.Provenance.Status == provenanceStatusCommitted || rejected.Provenance.Error == nil || rejected.Provenance.Error.Code != "provenance.conflict" {
		t.Fatalf("rejected check/provenance status = %#v", rejected)
	}
	if after := mustReadProvenanceFile(t, storePath); !bytes.Equal(after, before) {
		t.Fatal("conflicting check overwrote the ledger")
	}
	snapshot, err := provenance.New(storePath).Read(provenance.ReadOptions{})
	if err != nil || len(snapshot.Records) != 1 || snapshot.Records[0].Status != provenance.StatusCandidate {
		t.Fatalf("candidate evidence was promoted or lost: %#v, err=%v", snapshot, err)
	}
}
