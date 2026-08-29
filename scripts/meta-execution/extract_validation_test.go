package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

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
	unknown := validateBackupCleanup(backupCleanupObservation{Status: "PENDING"}, replacements)
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
