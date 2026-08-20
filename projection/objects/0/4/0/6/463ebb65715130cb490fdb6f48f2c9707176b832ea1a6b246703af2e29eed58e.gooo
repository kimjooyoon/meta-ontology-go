package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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
