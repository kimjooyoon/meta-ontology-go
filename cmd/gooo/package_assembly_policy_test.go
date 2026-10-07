package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const policyTarget = "../../examples/package-body-calls/gooo.workspace.json"
const policyCases = "../../examples/package-body-calls/cases.json"

func executePackagePolicyFixture(t *testing.T, policy string) packageExecutionReceipt {
	t.Helper()
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", policyCases,
		"--assembly-policy-workspace", policy, policyTarget}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	return receipt
}

func workspacePolicyRecord(t *testing.T, result *workspaceexecution.Result) *bodycodegen.RecordAssemblyReceipt {
	t.Helper()
	for i := range result.Composition.Steps {
		if record := result.Composition.Steps[i].Generation.Report.RecordAssembly; record != nil {
			return record
		}
	}
	t.Fatal("missing record construction")
	return nil
}

func replayPackagePolicyFixture(t *testing.T, receipt packageExecutionReceipt, wantSuccess bool) {
	t.Helper()
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "replay", "--json", "--receipt", path, "--cases", policyCases, policyTarget}, &out, &diagnostics)
	if (code == exitOK) != wantSuccess {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	if wantSuccess {
		var replay packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &replay); err != nil || replay.Result == nil ||
			replay.Result.Replay.ModelCalls != 0 || replay.Result.Runtime.FinitePassed != receipt.Result.Runtime.FinitePassed ||
			replay.Result.AssemblyPolicy == nil {
			t.Fatal("policy snapshot did not replay without inference", err, out.String())
		}
	}
}

func TestPackageAssemblyPolicyImportedHelpersAndPortableReplay(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"gooo.workspace.json", "counts.gooo.fixture", "policy.gooo.fixture"} {
		source, err := os.ReadFile(filepath.Join("../../examples/package-assembly-policy", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), source, 0600); err != nil {
			t.Fatal(err)
		}
	}
	receipt := executePackagePolicyFixture(t, filepath.Join(root, "gooo.workspace.json"))
	r := receipt.Result
	record := workspacePolicyRecord(t, r)
	if r.Runtime.FinitePassed != 8 || record.Passed != 5 || len(record.Attempts) != 4 ||
		record.Control == nil || len(record.Control.Decisions) != 4 ||
		r.AssemblyPolicy == nil || len(r.AssemblyPolicy.Program.PureCalls.Sites) != 1 {
		t.Fatal("imported Gooo policy did not control package construction", record)
	}
	if record.Control.Decisions[0].Operation != "CONTINUE_CANDIDATES" ||
		record.Control.Decisions[3].Operation != "OBSERVE_NEW_INPUTS" {
		t.Fatal("unexpected source policy decisions", record.Control.Decisions)
	}
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	replayPackagePolicyFixture(t, receipt, true)
}

func TestPackageAssemblyPolicyCanStopWithPartialEvidence(t *testing.T) {
	receipt := executePackagePolicyFixture(t, "../../examples/package-assembly-policy/checkpoint.workspace.json")
	record := workspacePolicyRecord(t, receipt.Result)
	if len(record.Attempts) != 1 || record.Passed >= record.Total || receipt.Result.Runtime.FinitePassed >= 8 ||
		record.Control.Decisions[0].Continue || receipt.Decision != "PROGRESS" {
		t.Fatal("policy stop lost incomplete observations", receipt.Decision, record)
	}
	replayPackagePolicyFixture(t, receipt, true)
}

func TestPackageAssemblyPolicyDecisionsMatchNativeProjection(t *testing.T) {
	policy := "../../examples/package-assembly-policy/gooo.workspace.json"
	receipt := executePackagePolicyFixture(t, policy)
	record := workspacePolicyRecord(t, receipt.Result)
	var rows []map[string]any
	for _, decision := range record.Control.Decisions {
		rows = append(rows, map[string]any{
			"inputs": map[string]any{"tools/assembly-policy:Explain": decision.Input},
			"expected": map[string]any{"tools/assembly-policy:Explain": map[string]string{
				"state": decision.State, "next_operation": decision.Operation, "message": decision.Message}},
		})
	}
	raw, err := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "policy-cases.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", path, policy}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var native packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &native); err != nil || native.Result.Runtime.FinitePassed != len(rows) {
		t.Fatal("policy evaluator and native package projection differ", err, out.String())
	}
}

func TestPackageAssemblyPolicyRejectsChangedSnapshotOrDecision(t *testing.T) {
	prior := executePackagePolicyFixture(t, "../../examples/package-assembly-policy/gooo.workspace.json")
	raw, _ := json.Marshal(prior)
	mutations := map[string]func(*workspaceexecution.Result){
		"missing packages": func(r *workspaceexecution.Result) { r.AssemblyPolicy = nil },
		"schema":           func(r *workspaceexecution.Result) { r.AssemblyPolicy.Schema = "other" },
		"source helper":    func(r *workspaceexecution.Result) { r.AssemblyPolicy.Manifest.Packages[1].Sources[0].Content += "\n" },
		"lowering":         func(r *workspaceexecution.Result) { r.AssemblyPolicy.Program.Source += "\n" },
		"control source":   func(r *workspaceexecution.Result) { workspacePolicyRecord(t, r).Control.Policy.Source += "\n" },
		"missing control":  func(r *workspaceexecution.Result) { workspacePolicyRecord(t, r).Control = nil },
		"decision":         func(r *workspaceexecution.Result) { workspacePolicyRecord(t, r).Control.Decisions[0].Continue = false },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var receipt packageExecutionReceipt
			if err := json.Unmarshal(raw, &receipt); err != nil {
				t.Fatal(err)
			}
			mutate(receipt.Result)
			replayPackagePolicyFixture(t, receipt, false)
		})
	}
}

func TestPackageAssemblyPolicyValidatesBeforeLoadingChoiceModel(t *testing.T) {
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", policyCases,
		"--assembly-policy-workspace", policyTarget, "--assembly-model", "/missing-choice-model", policyTarget}, &out, &diagnostics)
	if code != exitFailure || !strings.Contains(out.String(), "assembly policy requires one entry") {
		t.Fatal("invalid policy reached model loading", code, out.String(), diagnostics.String())
	}
}
