package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"os"
	"testing"
)

func TestRunProvenancePublishBindsAndRereadsCanonicalLedger(t *testing.T) {
	fixture := newProvenanceCLIFixture(t)
	beforeSource := mustReadProvenanceFile(t, fixture.sourcePath)
	output, code, stderr := fixture.publish(t, fixture.records)
	if code != exitOK || stderr != "" {
		t.Fatalf("publication = code %d, stderr=%q, output=%s", code, stderr, output)
	}
	response := decodeProvenanceResponse(t, output)
	if response.Schema != provenanceCLISchema || response.Status != provenanceStatusCommitted || response.Error != nil {
		t.Fatalf("publication response = %#v", response)
	}
	if response.SourceDigest != fixture.sourceDigest || response.SemanticDigest != fixture.semanticDigest || response.GraphDigest != fixture.graphDigest || response.StoreDigest == "" {
		t.Fatalf("publication identity = %#v", response)
	}
	if len(response.Records) != len(fixture.records) || response.Records[0].SemanticID != fixture.records[0].SemanticID {
		t.Fatalf("publication records = %#v", response.Records)
	}
	assertCanonicalProvenanceResponse(t, response)
	if got := mustReadProvenanceFile(t, fixture.sourcePath); !bytes.Equal(got, beforeSource) {
		t.Fatal("publication changed authoritative source")
	}
	if _, err := os.Stat(fixture.storePath + ".manifest"); err != nil {
		t.Fatalf("publication did not create committed manifest: %v", err)
	}
	store := provenance.New(fixture.storePath)
	snapshot, err := store.Read(provenance.ReadOptions{ExpectedSourceDigest: fixture.sourceDigest})
	if err != nil {
		t.Fatalf("canonical store reread = %v", err)
	}
	if snapshot.Digest != response.StoreDigest || len(snapshot.Records) != len(fixture.records) {
		t.Fatalf("store snapshot = %#v, response = %#v", snapshot, response)
	}
}
func TestRunProvenancePublishIdenticalReplayIsCanonicalAndOrderIndependent(t *testing.T) {
	fixture := newProvenanceCLIFixture(t)
	first, firstCode, firstStderr := fixture.publish(t, fixture.records)
	if firstCode != exitOK || firstStderr != "" {
		t.Fatalf("first publication = code %d, stderr=%q", firstCode, firstStderr)
	}
	replay := append([]provenance.Evidence(nil), fixture.records...)
	for left, right := 0, len(replay)-1; left < right; left, right = left+1, right-1 {
		replay[left], replay[right] = replay[right], replay[left]
	}
	second, secondCode, secondStderr := fixture.publish(t, replay)
	if secondCode != exitOK || secondStderr != "" {
		t.Fatalf("replay publication = code %d, stderr=%q, output=%s", secondCode, secondStderr, second)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("identical publication changed canonical output:\nfirst=%s\nsecond=%s", first, second)
	}
}
