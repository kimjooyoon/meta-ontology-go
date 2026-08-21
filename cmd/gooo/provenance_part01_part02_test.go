package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"testing"
)

func TestRunProvenancePublishConflictFailsClosedWithoutOverwrite(t *testing.T) {
	fixture := newProvenanceCLIFixture(t)
	if output, code, stderr := fixture.publish(t, fixture.records); code != exitOK || stderr != "" {
		t.Fatalf("setup publication = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	before := mustReadProvenanceFile(t, fixture.storePath)
	conflict := append([]provenance.Evidence(nil), fixture.records...)
	conflict[0].Attributes = map[string]string{"fixture": "conflict", "status": "verified"}
	output, code, stderr := fixture.publish(t, conflict)
	if code != exitFailure || stderr != "" {
		t.Fatalf("conflict publication = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	response := decodeProvenanceResponse(t, output)
	if response.Status == provenanceStatusCommitted || response.Error == nil || response.Error.Code != "provenance.conflict" {
		t.Fatalf("conflict response = %#v", response)
	}
	if after := mustReadProvenanceFile(t, fixture.storePath); !bytes.Equal(after, before) {
		t.Fatal("conflicting publication overwrote the committed ledger")
	}
}
