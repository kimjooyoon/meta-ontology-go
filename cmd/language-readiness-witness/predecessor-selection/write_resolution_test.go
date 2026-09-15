package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
)

func TestWriteResolutionFailurePreservesBlockedReceipt(t *testing.T) {
	report := predecessorresolution.Report{
		Decision: "FAIL_CLOSED", Reason: "READINESS_ANCESTOR_EVIDENCE_BLOCKED",
		ReportDigest: "sha256:blocked",
	}
	path := filepath.Join(t.TempDir(), "blocked.json")
	if err := writeResolutionFailure(path, report); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoded := predecessorresolution.Report{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Decision != report.Decision || decoded.Reason != report.Reason ||
		decoded.ReportDigest != report.ReportDigest {
		t.Fatalf("receipt = %#v", decoded)
	}
}
