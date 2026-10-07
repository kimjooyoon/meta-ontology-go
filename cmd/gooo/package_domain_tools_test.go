package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func executeDomainTool(t *testing.T, domain, mode string) packageExecutionReceipt {
	t.Helper()
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "domain-tools")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--" + mode,
		filepath.Join(root, domain+"."+mode+".json"), filepath.Join(root, domain+".workspace.json")}, &stdout, &stderr)
	if code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err, stdout.String())
	}
	return receipt
}

func TestDomainToolsShareReportAndRendererAcrossDistinctInputTypes(t *testing.T) {
	assembly := executeDomainTool(t, "assembly", "cases")
	documentation := executeDomainTool(t, "documentation", "cases")
	for _, receipt := range []packageExecutionReceipt{assembly, documentation} {
		runtime := receipt.Result.Runtime
		if runtime.FinitePassed != 14 || runtime.FiniteTotal != 14 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
			t.Fatal("domain tool did not satisfy its finite obligations", runtime)
		}
		if runtime.InputSeparation.DisjointInputs != 7 || runtime.InputSeparation.DisjointCasesPassed != 7 {
			t.Fatal("construction and evaluation inputs overlapped", runtime.InputSeparation)
		}
	}
	a, d := assembly.Result.Program.Workspace, documentation.Result.Program.Workspace
	if a.Digest == d.Digest {
		t.Fatal("domain replacement lost its different semantic identity")
	}
	for _, path := range []string{"gooo/tools/report", "gooo/tools/render"} {
		left, right := "", ""
		for _, pkg := range a.Packages {
			if pkg.Path == path {
				left = pkg.Sources[0].SourceDigest
			}
		}
		for _, pkg := range d.Packages {
			if pkg.Path == path {
				right = pkg.Sources[0].SourceDigest
			}
		}
		if left == "" || left != right {
			t.Fatal("shared contract or renderer changed", path, left, right)
		}
	}
}

func TestDomainToolReplacementRequiresItsOwnInputContract(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "domain-tools")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases",
		filepath.Join(root, "assembly.cases.json"), filepath.Join(root, "documentation.workspace.json")}, &stdout, &stderr)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err, stdout.String())
	}
	if code != exitFailure || receipt.Decision != "FAIL_CLOSED" || receipt.Error == "" {
		t.Fatal("domain-specific inputs were silently reinterpreted", code, receipt)
	}
}

func TestDomainToolsAcceptInputsWithoutAnExpectedAnswer(t *testing.T) {
	for _, domain := range []string{"assembly", "documentation"} {
		receipt := executeDomainTool(t, domain, "inputs")
		if receipt.Decision != "OBSERVED" || receipt.Result.Runtime.FiniteTotal != 0 || len(receipt.Result.Runtime.Traces) != 2 {
			t.Fatal("unscored input observations became a correctness claim", domain, receipt)
		}
	}
}
