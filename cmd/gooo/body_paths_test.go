package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func TestTypedPathCLIIsLocalWithConfiguredExternalProvider(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("GOOO_LAYA_URL", server.URL)
	t.Setenv("GOOO_LAYA_API_KEY", "fixture-key")
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	reader := mapSourceReader{"fixture.gooo": source, "plan.json": plan}
	args := []string{"--json", "--path-plan", "plan.json", "--activity", "Combined", "fixture.gooo"}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(args, reader, &stdout, &stderr)
	var result bodycodegen.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("JSON result: %v: %s", err, stdout.String())
	}
	if code != exitOK || stderr.Len() != 0 || result.Report.BodyPaths == nil || calls.Load() != 0 ||
		result.Report.BodyPaths.Search.Selection.ModelCalls != 0 || result.Report.RepositoryWrites != 0 {
		t.Fatalf("local CLI: code=%d calls=%d result=%+v stderr=%s", code, calls.Load(), result, stderr.String())
	}
	for _, replacement := range []string{
		strings.Replace(string(plan), `"expected":-16`, `"expected":null`, 1),
		strings.Replace(string(plan), `"max_attempts":8`, `"max_attempts":8,"max_attempts":1`, 1),
		strings.Replace(string(plan), `"input":-3,"expected":-16`, `"input":-3`, 1),
	} {
		reader["plan.json"] = []byte(replacement)
		stdout.Reset()
		stderr.Reset()
		code := runBodyCodegen(args, reader, &stdout, &stderr)
		if code != exitFailure || !strings.Contains(stdout.String(), `"decision":"FAIL_CLOSED"`) || calls.Load() != 0 {
			t.Fatalf("invalid typed plan was accepted: code=%d stdout=%s", code, stdout.String())
		}
	}
}

func TestTypedPathCLIRejectsAmbiguousModes(t *testing.T) {
	for _, flags := range [][]string{
		{"--path-model", "model.json"},
		{"--path-plan", "plan.json", "--fill-plan", "fill.json"},
		{"--path-plan", "plan.json", "--fill-search", "fill.json"},
		{"--path-plan", "plan.json", "--sample-seed", "seed"},
		{"--path-plan", "plan.json", "--tiny-model", "model.json"},
		{"--path-plan", "plan.json", "--path-plan", "again.json"},
		{"--path-step-attempts", "8"},
		{"--path-plan", "plan.json", "--path-step-attempts", "0"},
		{"--path-plan", "plan.json", "--path-step-attempts", "65"},
		{"--path-plan", "plan.json", "--path-step-attempts", "NaN"},
		{"--path-plan", "plan.json", "--path-step-attempts", "1", "--path-step-attempts", "2"},
	} {
		args := append(flags, "--activity", "Combined", "fixture.gooo")
		var stdout, stderr bytes.Buffer
		if code := runBodyCodegen(args, mapSourceReader{}, &stdout, &stderr); code != exitUsage {
			t.Fatalf("ambiguous mode accepted: %v: %d %s", flags, code, stderr.String())
		}
	}
}

func TestTypedPathCLIBatchesEmitLinkedPartialProgress(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment-ko-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	plan = bytes.Replace(plan, []byte(`"expected": 24`), []byte(`"expected": 999`), 1)
	reader := mapSourceReader{"fixture.gooo": source, "plan.json": plan}
	args := []string{"--json", "--path-plan", "plan.json", "--path-step-attempts", "8",
		"--activity", "ConditionalAssign", "fixture.gooo"}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(args, reader, &stdout, &stderr)
	var result bodycodegen.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err, stdout.String())
	}
	if code != exitOK || stderr.Len() != 0 || result.Report.BodyPaths == nil {
		t.Fatalf("batched CLI failed: %d %s", code, stderr.String())
	}
	receipt := result.Report.BodyPaths
	if receipt.Search.Status != "PARTIAL" || receipt.Search.Evaluated != 64 || len(receipt.Progress) != 9 ||
		receipt.FunctionalCompleteness != 600.0/7 || receipt.Progress[0].Attempted != 0 ||
		receipt.Progress[8].Attempted != 64 || receipt.Progress[8].Selection.ModelCalls != 0 {
		t.Fatalf("CLI lost incremental partial observations: %+v", receipt)
	}
}

func TestTypedPathCLIFailureUsesCommonPathReceipt(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	source = bytes.Replace(source, []byte("input + 2"), []byte("input + 99"), 1)
	args := []string{"--json", "--path-plan", "plan.json", "--path-model", "missing-model.json",
		"--activity", "Combined", "fixture.gooo"}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(args, mapSourceReader{"fixture.gooo": source, "plan.json": plan}, &stdout, &stderr)
	var result struct {
		Receipt json.RawMessage              `json:"completeness_receipt"`
		Paths   *bodycodegen.BodyPathReceipt `json:"body_paths"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	receipt, err := completeness.Decode(result.Receipt)
	if err != nil || code != exitFailure || receipt.ProfileID != "gooo/body-codegen-typed-path-v1" ||
		result.Paths == nil || result.Paths.SearchStarted || result.Paths.DocumentSHA256 == "" {
		t.Fatalf("failure lost path provenance: code=%d error=%v output=%s", code, err, stdout.String())
	}
	for _, d := range receipt.Dimensions {
		if d.ID == "typed_path_finite_accuracy" && (d.Status != "UNKNOWN" || d.Denominator != 3) {
			t.Fatalf("unobserved score: %+v", d)
		}
	}
}
