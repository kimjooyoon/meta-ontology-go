package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckSemanticProvenancePublishesAndReplaysCanonically(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "main.gooo")
	storePath := filepath.Join(directory, "ledger.jsonl")
	source := []byte(validSource)
	if err := os.WriteFile(sourcePath, source, 0o640); err != nil {
		t.Fatal(err)
	}
	beforeSource := append([]byte(nil), source...)

	first, firstCode, firstStderr := runCheckProvenanceJSON(t, sourcePath, storePath)
	if firstCode != exitOK || firstStderr != "" {
		t.Fatalf("first check = code %d, stderr=%q, output=%s", firstCode, firstStderr, first)
	}
	firstReport := decodeCheckJSON(t, first)
	assertCommittedCheckProvenance(t, firstReport)
	if len(firstReport.Provenance.Records) != 1 {
		t.Fatalf("first records = %#v", firstReport.Provenance.Records)
	}
	record := firstReport.Provenance.Records[0]
	if record.Kind != string(provenance.KindCompilerRun) || record.Status != string(provenance.StatusVerified) || record.Producer != string(semantic.GoHostedCompilerID) || record.SemanticID != record.ID {
		t.Fatalf("compiler-check record = %#v", record)
	}
	assertCanonicalProvenanceResponse(t, *firstReport.Provenance)
	ledgerBefore := mustReadProvenanceFile(t, storePath)

	second, secondCode, secondStderr := runCheckProvenanceJSON(t, sourcePath, storePath)
	if secondCode != exitOK || secondStderr != "" || !bytes.Equal(first, second) {
		t.Fatalf("replay = code %d, stderr=%q, output changed=%v\nfirst=%s\nsecond=%s", secondCode, secondStderr, !bytes.Equal(first, second), first, second)
	}
	if after := mustReadProvenanceFile(t, storePath); !bytes.Equal(after, ledgerBefore) {
		t.Fatal("identical replay changed the canonical ledger")
	}
	if err := os.WriteFile(storePath, []byte("divergent\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	repaired, repairedCode, repairedStderr := runCheckProvenanceJSON(t, sourcePath, storePath)
	if repairedCode != exitOK || repairedStderr != "" || !bytes.Equal(repaired, first) {
		t.Fatalf("committed repair = code %d, stderr=%q, output changed=%v", repairedCode, repairedStderr, !bytes.Equal(repaired, first))
	}
	if after := mustReadProvenanceFile(t, storePath); !bytes.Equal(after, ledgerBefore) {
		t.Fatal("check boundary did not preserve the committed ledger during repair")
	}
	if after := mustReadProvenanceFile(t, sourcePath); !bytes.Equal(after, beforeSource) {
		t.Fatal("semantic provenance check changed the authoritative source")
	}
	snapshot, err := provenance.New(storePath).Read(provenance.ReadOptions{ExpectedSourceDigest: firstReport.Provenance.SourceDigest})
	if err != nil || len(snapshot.Records) != 1 || snapshot.Digest != firstReport.Provenance.StoreDigest {
		t.Fatalf("canonical reread = %#v, err=%v", snapshot, err)
	}
}
