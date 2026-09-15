package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesReplayVerifiedReceiptOutsideRoot(t *testing.T) {
	root := t.TempDir()
	input := writeResolutionFixture(t, root)
	report := filepath.Join(t.TempDir(), "receipt.json")
	failClosed, err := run(config{root: root, input: input, report: report, check: true})
	if err != nil {
		t.Fatal(err)
	}
	if failClosed {
		t.Fatal("lower-resolution transition was rejected")
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var output receipt
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatal(err)
	}
	if !output.ReplayVerified || output.Report.Decision != "LOWER_RESOLUTION" ||
		output.RepositoryWrites != 0 || output.ReceiptDigest == "" {
		t.Fatalf("receipt = %#v", output)
	}
}

func TestRunRejectsReceiptInsideRepository(t *testing.T) {
	root := t.TempDir()
	input := writeResolutionFixture(t, root)
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
	input := writeResolutionFixture(t, root)
	report := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(report, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(config{root: root, input: input, report: report}); err == nil {
		t.Fatal("existing receipt was overwritten")
	}
}
