package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func newProvenanceCLIFixture(t *testing.T) provenanceCLIFixture {
	t.Helper()
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "main.gooo")
	evidencePath := filepath.Join(directory, "evidence.json")
	storePath := filepath.Join(directory, "provenance.jsonl")
	source := []byte(`package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`)
	if err := os.WriteFile(sourcePath, source, 0o640); err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() {
		t.Fatalf("fixture parse = %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatalf("fixture lower = %v", err)
	}
	sourceDigest := sha256Digest(source)
	produced := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	records := []provenance.Evidence{
		{
			ID: "billing://event/order-created", SemanticID: "billing://entity/order", Producer: "gooo://producer/cli-fixture",
			Kind: provenance.KindObservation, Status: provenance.StatusVerified,
			SourceSpan:   provenance.SourceSpan{URI: sourcePath, Start: provenance.Position{Offset: 0, Line: 1, Column: 1}, End: provenance.Position{Offset: len(source), Line: 10, Column: 1}},
			SourceDigest: sourceDigest, SemanticDigest: ir.StableHash(), GraphDigest: ir.Graph.StableHash(),
			Freshness:  provenance.NewFreshness(sourceDigest, produced, produced.Add(2*time.Hour)),
			Attributes: map[string]string{"fixture": "billing", "status": "verified"},
		},
		{
			ID: "billing://event/payment-created", SemanticID: "billing://entity/payment", Producer: "gooo://producer/cli-fixture",
			Kind: provenance.KindObservation, Status: provenance.StatusVerified,
			SourceSpan:   provenance.SourceSpan{URI: sourcePath, Start: provenance.Position{Offset: 0, Line: 1, Column: 1}, End: provenance.Position{Offset: len(source), Line: 10, Column: 1}},
			SourceDigest: sourceDigest, SemanticDigest: ir.StableHash(), GraphDigest: ir.Graph.StableHash(),
			Freshness:  provenance.NewFreshness(sourceDigest, produced, produced.Add(2*time.Hour)),
			Attributes: map[string]string{"fixture": "billing", "status": "verified"},
		},
	}
	return provenanceCLIFixture{
		directory: directory, sourcePath: sourcePath, evidencePath: evidencePath, storePath: storePath,
		sourceDigest: sourceDigest, semanticDigest: ir.StableHash(), graphDigest: ir.Graph.StableHash(), records: records,
	}
}

func (fixture provenanceCLIFixture) publish(t *testing.T, records []provenance.Evidence) ([]byte, int, string) {
	t.Helper()
	data, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.evidencePath, data, 0o640); err != nil {
		t.Fatal(err)
	}
	return fixture.publishRaw(t)
}

func (fixture provenanceCLIFixture) publishRaw(t *testing.T) ([]byte, int, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run([]string{"provenance", "publish", "--json", fixture.sourcePath, "--store", fixture.storePath, "--evidence", fixture.evidencePath}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.String()
}

func decodeProvenanceResponse(t *testing.T, data []byte) provenancePublishResponse {
	t.Helper()
	var response provenancePublishResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("decode provenance response %q: %v", data, err)
	}
	return response
}

func assertCanonicalProvenanceResponse(t *testing.T, response provenancePublishResponse) {
	t.Helper()
	want := response.CanonicalDigest
	response.CanonicalDigest = ""
	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Digest(payload); got != want {
		t.Fatalf("canonical response digest = %s, want %s", got, want)
	}
}

func decodeManifestFixture(t *testing.T, path string) map[string]any {
	t.Helper()
	data := mustReadProvenanceFile(t, path+".manifest")
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func mustReadProvenanceFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
