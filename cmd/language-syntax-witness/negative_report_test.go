package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestPersistValidatedNegativeReportKeepsFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	report := languagesyntax.Report{Decision: languagesyntax.DecisionClosed, Reason: "SYNTAX_ROUNDTRIP_MISMATCH"}
	negative := validatedNegativeReportError{decision: report.Decision, reason: report.Reason}
	var stdout bytes.Buffer
	err := persistBuiltReport(path, report, negative, &stdout)
	if err != negative || stdout.Len() != 0 {
		t.Fatalf("negative result became success: err=%v stdout=%s", err, &stdout)
	}
	raw, readErr := os.ReadFile(path)
	if readErr != nil || !bytes.Contains(raw, []byte(report.Reason)) {
		t.Fatalf("negative evidence missing: %s err=%v", raw, readErr)
	}
}

func TestInvalidReportDoesNotReplaceEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte("previous"), 0600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("validation failed")
	var stdout bytes.Buffer
	if err := persistBuiltReport(path, languagesyntax.Report{}, failure, &stdout); err != failure {
		t.Fatalf("validation failure changed: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "previous" || stdout.Len() != 0 {
		t.Fatalf("invalid evidence wrote output: %s err=%v", raw, err)
	}
}

func TestNegativeReportWriteFailureStaysFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "report.json")
	var stdout bytes.Buffer
	negative := validatedNegativeReportError{decision: "FAIL_CLOSED", reason: "fixture"}
	if err := persistBuiltReport(path, languagesyntax.Report{}, negative, &stdout); err == nil || err == negative || stdout.Len() != 0 {
		t.Fatalf("write failure hidden: err=%v stdout=%s", err, &stdout)
	}
}
