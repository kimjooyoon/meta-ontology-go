package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageStringOperationsExecuteAndReplay(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	manifest, cases, total := stringOpsWorkspace(t)
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", cases, manifest}, &output, &diagnostics); code != exitOK {
		t.Fatal("string operations failed", code, diagnostics.String(), output.String())
	}
	var original packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &original); err != nil {
		t.Fatal(err)
	}
	assertStringOpsRuntime(t, original, total)
	digest := sha256.Sum256(output.Bytes())
	receipt := filepath.Join(t.TempDir(), "execution.json")
	if err := os.WriteFile(receipt, output.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if code := run([]string{"package", "replay", "--json", "--receipt", receipt, "--cases", cases, manifest}, &output, &diagnostics); code != exitOK {
		t.Fatal("saved string operations failed", code, diagnostics.String(), output.String())
	}
	var replay packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &replay); err != nil {
		t.Fatal(err)
	}
	assertStringOpsRuntime(t, replay, total)
	if replay.ReplayedFrom != "sha256:"+hex.EncodeToString(digest[:]) || replay.Result.Replay == nil ||
		replay.Result.Replay.ModelCalls != 0 || replay.Result.Runtime.GeneratedSHA256 != original.Result.Runtime.GeneratedSHA256 {
		t.Fatal("saved replay changed its program or original receipt binding", replay)
	}
}

func assertStringOpsRuntime(t *testing.T, receipt packageExecutionReceipt, total int) {
	t.Helper()
	if receipt.Decision != "PASS" || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != total ||
		receipt.Result.Runtime.FiniteTotal != total || receipt.Result.Runtime.ModelCalls != 0 ||
		!receipt.Result.Runtime.RuntimeReplayed || !receipt.Result.Runtime.ProjectionReplayed ||
		receipt.Result.Program.Entry.PackagePath != "app/stringops" || receipt.Result.Program.Entry.Activity != "Main" {
		t.Fatal("native string operations did not retain finite results", receipt)
	}
}
