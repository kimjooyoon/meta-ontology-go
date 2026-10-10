package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCommitStagedPreservesPreexistingTemp(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "source.go")
	temporary := destination + ".extract.tmp"
	if err := os.WriteFile(destination, []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	owned := []byte("caller-owned temporary")
	if err := os.WriteFile(temporary, owned, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := commitStaged(map[string]stagedFile{
		"source.go": {name: destination, data: []byte("package p\n"), mode: 0o644},
	})
	if err == nil {
		t.Fatal("pre-existing temporary path was accepted")
	}
	got, err := os.ReadFile(temporary)
	if err != nil || !bytes.Equal(got, owned) {
		t.Fatalf("pre-existing temporary path was changed: %v", err)
	}
}

func TestTransactionFailurePreservesOriginalFiles(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "source.go")
	if err := os.WriteFile(destination, []byte("package old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	created := filepath.Join(root, "helper.go")
	transaction, err := commitStaged(map[string]stagedFile{
		"source.go": {name: destination, data: []byte("package new\n"), mode: 0o644},
		"helper.go": {name: created, data: []byte("package helper\n"), mode: 0o644, created: true},
	})
	if runtime.GOOS == "linux" {
		if err != nil {
			t.Fatal(err)
		}
		reportPath := filepath.Join(root, "report-directory")
		if err := os.Mkdir(reportPath, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := writeExtractionReport(reportPath, extractionReport{}); err == nil {
			t.Fatal("report publication unexpectedly succeeded")
		}
		if err := rollbackTransactions(transaction.files, len(transaction.files)); err != nil {
			t.Fatal(err)
		}
	} else if err == nil || !strings.Contains(err.Error(), "namespace replacement unsupported on GOOS "+runtime.GOOS) {
		t.Fatalf("unsupported host did not reject the transaction: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil || !bytes.Equal(got, []byte("package old\n")) {
		t.Fatalf("destination was not restored: %v", err)
	}
	if _, err := os.Lstat(created); !os.IsNotExist(err) {
		t.Fatalf("created helper survived rollback: %v", err)
	}
	for _, name := range []string{destination, created} {
		for _, suffix := range []string{".extract.tmp", ".extract.bak"} {
			if _, err := os.Lstat(name + suffix); !os.IsNotExist(err) {
				t.Fatalf("transaction residue %s: %v", name+suffix, err)
			}
		}
	}
}

func TestBackupCleanupFailureIsUnknown(t *testing.T) {
	result := removeTransactionBackups([]transactionFile{{backup: filepath.Join(t.TempDir(), "missing.bak"), destinationPreexisted: true, backupCreated: true}})
	if result.Status != "UNKNOWN" || result.Attempted != 1 || result.Removed != 0 || result.Failures != 1 {
		t.Fatalf("unexpected cleanup observation: %#v", result)
	}
}
