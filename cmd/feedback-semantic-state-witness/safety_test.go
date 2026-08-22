package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRejectsSemanticReceiptInsideRepository(t *testing.T) {
	root := t.TempDir()
	input, predecessor := writeSemanticFixture(t, root, "FIXED_POINT")
	output := filepath.Join(root, "semantic.json")
	if _, err := run(config{root: root, input: input,
		predecessorReceipt: predecessor, report: output}); err == nil {
		t.Fatal("semantic receipt inside repository was accepted")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("unexpected report state: %v", err)
	}
}

func TestRunDoesNotOverwriteSemanticReceipt(t *testing.T) {
	root := t.TempDir()
	input, predecessor := writeSemanticFixture(t, root, "FIXED_POINT")
	output := filepath.Join(t.TempDir(), "semantic.json")
	if err := os.WriteFile(output, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(config{root: root, input: input,
		predecessorReceipt: predecessor, report: output}); err == nil {
		t.Fatal("existing semantic receipt was overwritten")
	}
}
