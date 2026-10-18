package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageExecuteRetainsFailedNativeSetup(t *testing.T) {
	for _, compression := range []string{"", "gzip"} {
		t.Run("compression="+compression, func(t *testing.T) { testPackageExecuteFailedSetup(t, compression) })
	}
}

func testPackageExecuteFailedSetup(t *testing.T, compression string) {
	t.Helper()
	root := filepath.Join("..", "..", "examples", "package-body-calls")
	manifest, cases := filepath.Join(root, "source.workspace.json"), filepath.Join(root, "cases.json")
	var stdout, stderr bytes.Buffer
	args := []string{"package", "execute", "--json", "--cases", cases, "--go", filepath.Join(t.TempDir(), "missing-go"), manifest}
	if compression != "" {
		args = append(args, "--receipt-compression", compression)
	}
	code := run(args, &stdout, &stderr)
	raw, err := decodePackageReceiptTransport(stdout.Bytes(), packageReceiptMaxBytes)
	if err != nil {
		t.Fatal(err)
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if code != exitFailure || stderr.Len() != 0 || receipt.Decision != "FAIL_CLOSED" || receipt.Error == "" || receipt.Result == nil {
		t.Fatalf("failed execution lost its result: code=%d receipt=%+v stderr=%s", code, receipt, stderr.String())
	}
	r := receipt.Result
	if r.Schema == "" || r.Composition.GeneratedSHA256 == "" || r.Runtime.Stage == "" || r.Runtime.Failure == "" ||
		r.Runtime.FinitePassed != 0 || r.Runtime.FiniteTotal != 8 || r.Runtime.RuntimeReplayed || len(r.Runtime.Runs) != 0 {
		t.Fatalf("native setup evidence differs: %+v", r.Runtime)
	}
	path := filepath.Join(t.TempDir(), "failed.json")
	if err := os.WriteFile(path, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"package", "replay", "--json", "--receipt", path, "--cases", cases, manifest}, &stdout, &stderr)
	var rejected packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &rejected); err != nil {
		t.Fatal(err)
	}
	if code != exitFailure || !strings.Contains(stdout.String(), "saved execution envelope") ||
		rejected.Failure == nil || rejected.Failure.Code != "REPLAY_ENVELOPE_INVALID" ||
		rejected.Failure.Detail != rejected.Error || rejected.InputsDigest != "" || rejected.CasesDigest != "" {
		t.Fatalf("failed execution became replayable: code=%d stdout=%s", code, stdout.String())
	}
}
