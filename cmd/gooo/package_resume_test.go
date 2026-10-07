package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func runSavedPackageFixture(t *testing.T, args ...string) (packageExecutionReceipt, []byte) {
	t.Helper()
	var out, diagnostics bytes.Buffer
	if code := run(append([]string{"package"}, args...), &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err, out.String())
	}
	return receipt, append([]byte(nil), out.Bytes()...)
}

func TestPackageResumeContinuesImportedCalledHelper(t *testing.T) {
	root := "../../examples/called-body-construction/"
	policy := t.TempDir() + string(os.PathSeparator)
	for _, name := range []string{"gooo.workspace.json", "policy.gooo.fixture", "counts.gooo.fixture", "checkpoint.workspace.json", "checkpoint.gooo.fixture"} {
		raw, err := os.ReadFile(filepath.Join("../../examples/package-assembly-policy", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(policy, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	first, raw := runSavedPackageFixture(t, "execute", "--json", "--cases", root+"cases.json",
		"--assembly-policy-workspace", policy+"checkpoint.workspace.json", root+"gooo.workspace.json")
	if first.Result.Runtime.FinitePassed != 2 || first.Decision != "PROGRESS" {
		t.Fatal("checkpoint did not retain partial native results")
	}
	path := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "unexpected inference", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("GOOO_LAYA_URL", server.URL)
	t.Setenv("GOOO_LAYA_API_KEY", "unused-local-test")
	continued, raw := runSavedPackageFixture(t, "resume", "--json", "--receipt", path,
		"--cases", root+"cases.json", "--assembly-policy-workspace", policy+"gooo.workspace.json", root+"gooo.workspace.json")
	if continued.Result.Runtime.FinitePassed != 4 || continued.Decision != "PASS" ||
		continued.Result.Composition.Continuation == nil || continued.Result.Composition.Continuation.NewModelCalls != 0 {
		t.Fatal("package continuation did not finish its finite native expectations")
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(policy); err != nil {
		t.Fatal(err)
	}
	replay, _ := runSavedPackageFixture(t, "replay", "--json", "--receipt", path,
		"--cases", root+"cases.json", root+"gooo.workspace.json")
	if replay.Result.Runtime.FinitePassed != 4 || replay.Result.Replay.ModelCalls != 0 || calls.Load() != 0 ||
		continued.ContinuedFrom == "" || continued.ReplayedFrom != "" || len(replay.Result.PolicyHistory) != 1 ||
		replay.Result.Continuation == nil || replay.Result.Runtime.InputSeparation.DisjointCasesPassed != 4 {
		t.Fatal("continued package did not replay without inference")
	}
}

func TestPackageResumeRequiresExplicitPolicyAndRejectsInferenceOptions(t *testing.T) {
	base := []string{"package", "resume", "--receipt", "saved.json", "--cases", "cases.json"}
	for _, extra := range [][]string{
		nil,
		{"--assembly-policy-workspace", "policy.json", "--assembly-model", "model.json"},
		{"--assembly-policy-workspace", "policy.json", "--tiny-model", "model.json"},
		{"--assembly-policy-workspace", "policy.json", "--body-plans", "plans.json"},
		{"--assembly-policy-workspace", "policy.json", "--inputs", "inputs.json"},
		{"--assembly-policy-workspace", "policy.json", "--assembly-policy-workspace", "other.json"},
	} {
		args := append(append(append([]string(nil), base...), extra...), "workspace.json")
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitUsage {
			t.Fatal("ambiguous or inference option accepted", args, code, out.String(), diagnostics.String())
		}
	}
}
