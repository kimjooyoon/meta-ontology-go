package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestRunRejectsNegativePredecessorRepositoryWritesWithoutReceipt(t *testing.T) {
	root := t.TempDir()
	input, predecessor := writeSemanticFixture(t, root, "FIXED_POINT")
	data, err := os.ReadFile(predecessor)
	if err != nil {
		t.Fatal(err)
	}
	var receipt predecessorReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	receipt.RepositoryWrites = -1
	data, err = json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(predecessor, data, 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "semantic.json")
	if _, err := run(config{root: root, input: input,
		predecessorReceipt: predecessor, report: output}); err == nil || !strings.Contains(err.Error(), "NEGATIVE_REPOSITORY_WRITES") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("unexpected report state: %v", err)
	}
}

func TestRunRejectsPredecessorRepositoryWriteOverflowWithoutReceipt(t *testing.T) {
	root := t.TempDir()
	input, predecessor := writeSemanticFixture(t, root, "FIXED_POINT")
	data, err := os.ReadFile(predecessor)
	if err != nil {
		t.Fatal(err)
	}
	var receipt predecessorReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	maxInt := int(^uint(0) >> 1)
	receipt.RepositoryWrites = maxInt
	receipt.Report.Summary.RepositoryWrites = 1
	data, err = json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(predecessor, data, 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "semantic.json")
	if _, err := run(config{root: root, input: input,
		predecessorReceipt: predecessor, report: output}); err == nil || !strings.Contains(err.Error(), "REPOSITORY_WRITES_OVERFLOW") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("unexpected report state: %v", err)
	}
}
