package main

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPackageExecuteRunsGoooAssemblyExplainer(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "assembly-explainer")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"),
		filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr)
	if code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	result := receipt.Result
	if receipt.Decision != "PASS" || result == nil || result.Runtime.FinitePassed != 10 || result.Runtime.FiniteTotal != 10 || !result.Runtime.RuntimeReplayed {
		t.Fatal("Gooo explainer did not produce all finite observations", receipt)
	}
	if len(result.Program.Activities) != 1 || len(result.Composition.Plan.Edges) != 0 || strings.Contains(result.Program.Source, "bind ") {
		t.Fatal("standalone tool contains an invented producer or binding")
	}
	if result.Composition.Plan.Activities[0].InputEntityID != "gooo://tools/assembly-observation" || result.Runtime.ModelCalls != 0 {
		t.Fatal("tool lost semantic identity or inferred a model")
	}
	if result.Runtime.InputSeparation.Status != "UNKNOWN" {
		t.Fatal("ordinary tool claimed a construction-disjoint search observation")
	}
}

func TestRunPackageExecuteInputOnlyReturnsActualValuesAndUnknownExpectations(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "assembly-explainer")
	args := []string{"package", "execute", "--inputs", filepath.Join(root, "inputs.json"), filepath.Join(root, "gooo.workspace.json")}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	decoder := json.NewDecoder(&stdout)
	for _, next := range []string{"USE_OBSERVED_CANDIDATE", "OBSERVE_NEW_INPUTS", "EVALUATE_CANDIDATES"} {
		var actual map[string]string
		if err := decoder.Decode(&actual); err != nil || actual["next_operation"] != next {
			t.Fatal("unexpected Gooo tool output", err, actual)
		}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatal("extra entry output", err)
	}
	stdout.Reset()
	if code := run(append(args, "--json"), &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "OBSERVED" || receipt.InputsDigest == "" || receipt.CasesDigest != "" || receipt.Result == nil {
		t.Fatal("input-only receipt identity differs", receipt)
	}
	r := receipt.Result.Runtime
	if !r.RuntimeReplayed || r.FiniteTotal != 0 || r.FinitePassed != 0 || r.InputSeparation.Status != "UNKNOWN" || r.InputSeparation.DisjointCasesPassed != 0 {
		t.Fatal("actual values became a correctness claim", r)
	}
	for _, trace := range r.Traces {
		for _, delivery := range trace.Deliveries {
			if delivery.Passed != nil || delivery.Expected != nil {
				t.Fatal("input-only run invented an expectation")
			}
		}
	}
}

func TestRunPackageExecuteRejectsCasesTogetherWithInputs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--cases", "unused", "--inputs", "unused", "unused"}, &stdout, &stderr); code != exitUsage {
		t.Fatal("ambiguous execution mode was accepted", code)
	}
}
