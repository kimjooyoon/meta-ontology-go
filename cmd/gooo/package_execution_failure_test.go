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
	root := filepath.Join("..", "..", "examples", "package-body-calls")
	manifest, cases := filepath.Join(root, "source.workspace.json"), filepath.Join(root, "cases.json")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", cases,
		"--go", filepath.Join(t.TempDir(), "missing-go"), manifest}, &stdout, &stderr)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
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
	if code != exitFailure || !strings.Contains(stdout.String(), "saved execution envelope") {
		t.Fatalf("failed execution became replayable: code=%d stdout=%s", code, stdout.String())
	}
}
