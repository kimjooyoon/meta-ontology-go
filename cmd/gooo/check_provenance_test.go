package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func TestRunCheckSemanticProvenanceDeclarationPermutationKeepsSemanticBinding(t *testing.T) {
	directory := t.TempDir()
	firstPath := filepath.Join(directory, "first.gooo")
	secondPath := filepath.Join(directory, "second.gooo")
	firstStore := filepath.Join(directory, "first.jsonl")
	secondStore := filepath.Join(directory, "second.jsonl")
	firstSource := validSource
	secondSource := `package billing
namespace billing
activity PayOrder(Order) -> Order
entity Order id "billing://entity/order"
`
	for path, source := range map[string]string{firstPath: firstSource, secondPath: secondSource} {
		if err := os.WriteFile(path, []byte(source), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	first, code, stderr := runCheckProvenanceJSON(t, firstPath, firstStore)
	if code != exitOK || stderr != "" {
		t.Fatalf("first permutation check = code %d, stderr=%q", code, stderr)
	}
	second, code, stderr := runCheckProvenanceJSON(t, secondPath, secondStore)
	if code != exitOK || stderr != "" {
		t.Fatalf("second permutation check = code %d, stderr=%q", code, stderr)
	}
	left, right := decodeCheckJSON(t, first), decodeCheckJSON(t, second)
	if left.Provenance.SemanticDigest != right.Provenance.SemanticDigest || left.Provenance.GraphDigest != right.Provenance.GraphDigest || left.Provenance.Records[0].Kind != right.Provenance.Records[0].Kind || left.Provenance.Records[0].Producer != right.Provenance.Records[0].Producer {
		t.Fatalf("permutation changed semantic evidence binding: left=%#v right=%#v", left.Provenance, right.Provenance)
	}
	if left.Provenance.SourceDigest == right.Provenance.SourceDigest || left.Provenance.Records[0].ID == right.Provenance.Records[0].ID {
		t.Fatal("permutation incorrectly erased source-bound event identity")
	}
}

func runCheckProvenanceJSON(t *testing.T, sourcePath, storePath string) ([]byte, int, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run([]string{"check", "--semantic", "--json", "--provenance-store", storePath, sourcePath}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.String()
}

func decodeCheckJSON(t *testing.T, data []byte) jsonReport {
	t.Helper()
	var report jsonReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode check JSON %q: %v", data, err)
	}
	return report
}

func assertCommittedCheckProvenance(t *testing.T, report jsonReport) {
	t.Helper()
	if report.Status != "ok" || report.Provenance == nil || report.Provenance.CheckStatus != checkStatusPass || report.Provenance.Status != provenanceStatusCommitted || report.Provenance.Error != nil || report.SemanticHash != report.Provenance.SemanticDigest || report.Provenance.StoreDigest == "" {
		t.Fatalf("check/provenance report = %#v", report)
	}
}
