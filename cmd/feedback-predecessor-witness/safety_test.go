package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRejectsReceiptInsideRepository(t *testing.T) {
	root := t.TempDir()
	input := writeFixture(t, root, false)
	report := filepath.Join(root, "receipt.json")
	if _, err := run(config{root: root, input: input, report: report}); err == nil {
		t.Fatal("receipt inside repository was accepted")
	}
	if _, err := os.Stat(report); !os.IsNotExist(err) {
		t.Fatalf("unexpected report state: %v", err)
	}
}

func TestRunDoesNotOverwriteReceipt(t *testing.T) {
	root := t.TempDir()
	input := writeFixture(t, root, false)
	report := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(report, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(config{root: root, input: input, report: report}); err == nil {
		t.Fatal("existing receipt was overwritten")
	}
}
