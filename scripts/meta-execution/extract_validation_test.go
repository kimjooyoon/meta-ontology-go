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
	return extractionTestIndicatorsWithValues(0, 0, 0, 0, 0)
}

func extractionTestIndicatorsWithValues(observed, staged, applied, created, unhandled int) []json.RawMessage {
	values := []extractorIndicatorRecord{
		{ID: "extraction.observed", Value: observed, Limit: -1, Consumer: "function-extractor", Operation: "observe-density-residual", Proof: "axiomatic-foundation"},
		{ID: "extraction.staged", Value: staged, Limit: -1, Consumer: "function-extractor", Operation: "stage-helper-extraction", Proof: "coherent-system"},
		{ID: "extraction.applied", Value: applied, Limit: -1, Consumer: "logical-materializer", Operation: "accept-helper-extraction", Proof: "coherent-system"},
		{ID: "extraction.created", Value: created, Limit: -1, Consumer: "authorized-write-set", Operation: "authorize-declared-file-creation", Proof: "axiomatic-foundation"},
		{ID: "extraction.unhandled", Value: unhandled, Limit: 0, Blocking: true, Consumer: "function-extractor", Operation: "define-extraction-recipe", Proof: "infinite-regress"},
	}
	result := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		encoded, _ := json.Marshal(value)
		result = append(result, encoded)
	}
	return result
}

func TestExtractorReportRejectsIndicatorAndAggregateDrift(t *testing.T) {
	indicators := extractionTestIndicators()
	indicators = append(indicators, indicators[0])
	if _, ok := extractorIndicatorValues(indicators); ok {
		t.Fatal("extra indicator was accepted")
	}
	var wrong extractorIndicatorRecord
	if err := json.Unmarshal(extractionTestIndicators()[0], &wrong); err != nil {
		t.Fatal(err)
	}
	wrong.Operation = "wrong-operation"
	encoded, _ := json.Marshal(wrong)
	if _, ok := extractorIndicatorValues([]json.RawMessage{encoded}); ok {
		t.Fatal("wrong indicator operation was accepted")
	}
	validIndicators := extractionTestIndicatorsWithValues(2, 2, 2, 2, 0)
	cases := []struct {
		name     string
		report   extractorReport
	}{
		{"duplicate-subject", extractorReport{StagedSubjects: 2, Subjects: []extractorSubject{
			{Logical: "a", State: "COMMITTED_APPLIED", Before: 1, After: 1, Files: []string{"a.go"}},
			{Logical: "a", State: "COMMITTED_APPLIED", Before: 1, After: 1, Files: []string{"b.go"}},
		}, Indicators: validIndicators}},
		{"duplicate-changed-file", extractorReport{StagedSubjects: 2, Subjects: []extractorSubject{
			{Logical: "a", State: "COMMITTED_APPLIED", Before: 1, After: 1, Files: []string{"same.go"}},
			{Logical: "b", State: "COMMITTED_APPLIED", Before: 1, After: 1, Files: []string{"same.go"}},
		}, Indicators: validIndicators}},
		{"duplicate-created-file", extractorReport{StagedSubjects: 2, Subjects: []extractorSubject{
			{Logical: "a", State: "COMMITTED_APPLIED", Before: 1, After: 1, Files: []string{"a.go"}, CreatedFiles: []string{"helper.go"}},
			{Logical: "b", State: "COMMITTED_APPLIED", Before: 1, After: 1, Files: []string{"b.go"}, CreatedFiles: []string{"helper.go"}},
		}, Indicators: validIndicators}},
		{"invalid-state", extractorReport{StagedSubjects: 1, Subjects: []extractorSubject{
			{Logical: "a", State: "UNKNOWN", Before: 1, After: 1, Files: []string{"a.go"}},
		}, Indicators: extractionTestIndicatorsWithValues(1, 1, 1, 0, 0)}},
		{"nonpositive-lines", extractorReport{StagedSubjects: 1, Subjects: []extractorSubject{
			{Logical: "a", State: "COMMITTED_APPLIED", Before: 0, After: 1, Files: []string{"a.go"}},
		}, Indicators: extractionTestIndicatorsWithValues(1, 1, 1, 0, 0)}},
	}
	for _, test := range cases {
		if !validateExtractorReport(test.report) {
			t.Errorf("%s aggregate contradiction was accepted", test.name)
		}
	}
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
