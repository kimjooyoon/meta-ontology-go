package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestValidationRejectsMissingImportFromOutputUnion(t *testing.T) {
	root := t.TempDir()
	before := []byte("package p\n\nimport \"fmt\"\n\nfunc Selected() { _ = fmt.Sprint(1) }\n")
	after := []byte("package p\n\nfunc Selected() {}\n")
	if err := os.WriteFile(filepath.Join(root, "subject.go"), after, 0o644); err != nil {
		t.Fatal(err)
	}
	set, file, err := parseGoFile("subject.go", before)
	if err != nil {
		t.Fatal(err)
	}
	subject := sourcepolicy.SourceSubject{Path: "subject.go", Name: "Selected"}
	observed := extractorSubject{Logical: subject.Path, Files: []string{subject.Path}}
	_, _, _, err = validateOutputFiles(root, subject, file.Name.Name, importIdentity(file), sourceHeader(before, set, file), observed)
	if err == nil {
		t.Fatal("output import union omission was accepted")
	}
}

func TestValidationSeparatesBackupUnknownFromContradiction(t *testing.T) {
	replacements := []namespaceReplacementReceipt{{DestinationPreexisted: true}}
	unknown := validateBackupCleanup(backupCleanupObservation{Status: "PENDING", Attempted: 1}, replacements)
	var unavailable *extractValidationUnknown
	if !errors.As(unknown, &unavailable) || unavailable.reason != "BACKUP_CLEANUP_UNAVAILABLE" {
		t.Fatalf("pending cleanup was not unknown: %v", unknown)
	}
	contradiction := validateBackupCleanup(backupCleanupObservation{Status: "PASS", Attempted: 1}, replacements)
	var replacementErr *namespaceReplacementError
	if !errors.As(contradiction, &replacementErr) || replacementErr.reason != "BACKUP_CLEANUP_INCONSISTENT" {
		t.Fatalf("inconsistent cleanup was not refuted: %v", contradiction)
	}
}

func TestValidationRejectsCleanupCountContradictionBeforeUnknown(t *testing.T) {
	replacements := []namespaceReplacementReceipt{{DestinationPreexisted: true}}
	for _, observation := range []backupCleanupObservation{
		{Status: "UNKNOWN", Attempted: -1, Removed: 0, Failures: 1},
		{Status: "UNKNOWN", Attempted: 1, Removed: 1, Failures: 1},
	} {
		err := validateBackupCleanup(observation, replacements)
		var replacementErr *namespaceReplacementError
		if !errors.As(err, &replacementErr) || replacementErr.reason != "BACKUP_CLEANUP_INCONSISTENT" {
			t.Fatalf("cleanup contradiction was not refuted: %v", err)
		}
	}
}

func TestFailedExtractionPreservesBackupUnknownCoordinate(t *testing.T) {
	root := t.TempDir()
	report := extractorReport{
		Schema: functionExtractionReportSchema, SourceSHA: "head",
		Indicators: extractionTestIndicators(),
		NamespaceReplacements: []namespaceReplacementReceipt{{DestinationPreexisted: true}},
		BackupCleanup: backupCleanupObservation{Status: "UNKNOWN", Attempted: 1, Failures: 1},
	}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "report.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	failure := failedExtractionError(root, "report.json", generation.Plan{HeadSHA: "head"})
	if failure.stage != "evaluate-operation" || failure.step != "validate-function-extraction" ||
		failure.reason != "BACKUP_CLEANUP_UNAVAILABLE" || failure.class != "DIRECT_MISSING" ||
		failure.next != "recover-backup-cleanup-evidence" {
		t.Fatalf("backup unknown coordinate was not preserved: %+v", failure)
	}
}

func extractionTestIndicators() []json.RawMessage {
	values := []struct {
		id    string
		value int
	}{
		{"extraction.observed", 0}, {"extraction.staged", 0},
		{"extraction.applied", 0}, {"extraction.created", 0}, {"extraction.unhandled", 0},
	}
	result := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		encoded, _ := json.Marshal(extractorIndicatorRecord{ID: value.id, Value: value.value})
		result = append(result, encoded)
	}
	return result
}

func TestFailedExtractionMalformedReportIsRefuted(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "report.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	failure := failedExtractionError(root, "report.json", generation.Plan{HeadSHA: "head"})
	if failure.stage != "evaluate-operation" || failure.step != "decode-function-extraction-report" ||
		failure.reason != "INSTANCE_EVIDENCE_MALFORMED" || failure.class != "KNOWN_CONTRADICTION" ||
		failure.next != "report-counterexample" {
		t.Fatalf("malformed report was not refuted: %+v", failure)
	}
}
