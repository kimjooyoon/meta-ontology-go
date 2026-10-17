package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageRSSMetadataExampleExecutesAndReplays(t *testing.T) {
	root := "../../examples/measurement-metadata"
	manifest, cases := filepath.Join(root, "gooo.workspace.json"), filepath.Join(root, "cases.json")
	receipt, raw := runSavedPackageFixture(t, "execute", "--json", "--cases", cases, manifest)
	// Each complete record is one expected output entry, with four authored fields.
	if receipt.Decision != "PASS" || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 14 ||
		receipt.Result.Runtime.FiniteTotal != 14 || len(receipt.Result.Runtime.Traces) != 14 ||
		!receipt.Result.Runtime.RuntimeReplayed || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("metadata example did not match fourteen four-field expectations", string(raw))
	}
	receiptPath := filepath.Join(t.TempDir(), "metadata-execution.json")
	if err := os.WriteFile(receiptPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	receipt, raw = runSavedPackageFixture(t, "replay", "--json", "--receipt", receiptPath, "--cases", cases, manifest)
	if receipt.Decision != "PASS" || receipt.Result.Replay == nil || receipt.ReplayedFrom == "" ||
		receipt.Result.Runtime.FinitePassed != 14 || receipt.Result.Runtime.FiniteTotal != 14 ||
		len(receipt.Result.Runtime.Traces) != 14 || !receipt.Result.Runtime.RuntimeReplayed ||
		receipt.Result.Runtime.ModelCalls != 0 || receipt.Result.Replay.ModelCalls != 0 {
		t.Fatal("saved explanation did not preserve its declared expectations", string(raw))
	}
}
