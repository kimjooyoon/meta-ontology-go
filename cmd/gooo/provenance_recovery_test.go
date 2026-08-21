package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestRunProvenancePublishRepairsCommittedAndPreparedLedgerState(t *testing.T) {
	fixture := newProvenanceCLIFixture(t)
	if output, code, stderr := fixture.publish(t, fixture.records); code != exitOK || stderr != "" {
		t.Fatalf("setup publication = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	expected := mustReadProvenanceFile(t, fixture.storePath)
	if err := os.WriteFile(fixture.storePath, []byte("divergent\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	output, code, stderr := fixture.publish(t, fixture.records)
	if code != exitOK || stderr != "" {
		t.Fatalf("committed repair publication = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	if after := mustReadProvenanceFile(t, fixture.storePath); !bytes.Equal(after, expected) {
		t.Fatal("CLI did not preserve the exact committed ledger during repair")
	}

	prepared := newProvenanceCLIFixture(t)
	base := []provenance.Evidence{prepared.records[0]}
	post := append([]provenance.Evidence(nil), base...)
	post = append(post, prepared.records[1])
	basePath := filepath.Join(prepared.directory, "base.jsonl")
	postPath := filepath.Join(prepared.directory, "post.jsonl")
	if err := provenance.New(basePath).Append(base...); err != nil {
		t.Fatal(err)
	}
	if err := provenance.New(postPath).Append(post...); err != nil {
		t.Fatal(err)
	}
	baseManifest := decodeManifestFixture(t, basePath)
	postManifest := decodeManifestFixture(t, postPath)
	preparedManifest := map[string]any{
		"schema": postManifest["schema"], "phase": "prepared", "bytes": postManifest["bytes"],
		"lines": postManifest["lines"], "digest": postManifest["digest"], "last_id": postManifest["last_id"],
		"last_hash": postManifest["last_hash"], "data": postManifest["data"],
		"base": map[string]any{
			"bytes": baseManifest["bytes"], "lines": baseManifest["lines"], "digest": baseManifest["digest"],
			"last_id": baseManifest["last_id"], "last_hash": baseManifest["last_hash"], "data": baseManifest["data"],
		},
	}
	if err := os.WriteFile(prepared.storePath, mustReadProvenanceFile(t, postPath), 0o640); err != nil {
		t.Fatal(err)
	}
	preparedBytes, err := json.Marshal(preparedManifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(prepared.storePath+".manifest", append(preparedBytes, '\n'), 0o640); err != nil {
		t.Fatal(err)
	}
	output, code, stderr = prepared.publish(t, base)
	if code != exitOK || stderr != "" {
		t.Fatalf("prepared recovery publication = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	response := decodeProvenanceResponse(t, output)
	if response.Status != provenanceStatusCommitted || len(response.Records) != 1 {
		t.Fatalf("prepared recovery response = %#v", response)
	}
	manifest := decodeManifestFixture(t, prepared.storePath)
	if manifest["phase"] != "committed" {
		t.Fatalf("prepared manifest was not recovered: %#v", manifest)
	}
}
