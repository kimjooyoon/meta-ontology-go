package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunSourceInputExecutesRuntimeBindingPlan(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve runtime binding fixture path")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "..", "..", "examples", "language-runtime-binding", "main.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	reader := runSourceReaderWithFiles{
		"examples/language-runtime-binding/main.gooo": source,
		"input.json": []byte(`{"value":41}`),
	}
	var stdout, stderr bytes.Buffer
	code := runSource([]string{"--json", "--entry", "Produce", "--input", "input.json", "examples/language-runtime-binding/main.gooo"}, reader, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var report struct {
		Decision  string `json:"decision"`
		Execution struct {
			ApplyCalls int `json:"apply_calls"`
			Deliveries int `json:"deliveries"`
			Results    map[string]struct {
				Value int64 `json:"value"`
			} `json:"results"`
		} `json:"execution"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PASS" || report.Execution.ApplyCalls != 3 || report.Execution.Deliveries != 2 ||
		report.Execution.Results["ConsumeA"].Value != 43 || report.Execution.Results["ConsumeB"].Value != 43 {
		t.Fatalf("CLI plan report = %#v", report)
	}
}

func TestRunSourceRejectsBoundUnknownAndUnrelatedRootEntries(t *testing.T) {
	fanoutSource := []byte(`package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity Produce(Integer) -> Integer computes "int.add:1"
activity ConsumeA(Integer) -> Integer computes "int.add:1"
activity ConsumeB(Integer) -> Integer computes "int.add:1"
bind Produce.result -> ConsumeA.input
bind Produce.result -> ConsumeB.input
`)
	rootFixture := []byte(`package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity Alpha(Integer) -> Integer computes "int.add:1"
activity Beta(Integer) -> Integer computes "int.add:1"
`)
	cases := []struct {
		name       string
		source     []byte
		entry      string
		wantReason string
		wantStep   string
	}{
		{name: "bound entry", source: fanoutSource, entry: "ConsumeA", wantReason: "VALUE_EXTERNAL_INPUT_UNEXPECTED", wantStep: "reject-bound-external-input"},
		{name: "unknown entry", source: fanoutSource, entry: "Missing", wantReason: "VALUE_EXTERNAL_INPUT_UNEXPECTED", wantStep: "reject-unknown-external-input"},
		{name: "unrelated root", source: rootFixture, entry: "Alpha", wantReason: "VALUE_EXTERNAL_INPUT_MISSING", wantStep: "require-external-root-input"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := runSourceReaderWithFiles{
				"fixture.gooo": testCase.source,
				"input.json":  []byte(`{"value":41}`),
			}
			var stdout, stderr bytes.Buffer
			code := runSource([]string{"--json", "--entry", testCase.entry, "--input", "input.json", "fixture.gooo"}, reader, &stdout, &stderr)
			if code != exitFailure || stderr.Len() != 0 {
				t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
			}
			var report struct {
				Decision string `json:"decision"`
				Reason   string `json:"reason"`
				Failure  struct {
					Stage string `json:"stage"`
					Step  string `json:"step"`
				} `json:"failure"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if report.Decision != "FAIL_CLOSED" || report.Reason != testCase.wantReason || report.Failure.Stage != "EXECUTE" || report.Failure.Step != testCase.wantStep {
				t.Fatalf("failure report = %#v, want reason=%s step=%s", report, testCase.wantReason, testCase.wantStep)
			}
		})
	}
}

func TestRunSourceFailurePreservesPartialExecutionEvidence(t *testing.T) {
	source := []byte(`package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity Produce(Integer) -> Integer computes "int.add:1"
`)
	reader := runSourceReaderWithFiles{
		"overflow.gooo": source,
		"input.json":    []byte(`{"value":9223372036854775807}`),
	}
	var stdout, stderr bytes.Buffer
	code := runSource([]string{"--json", "--entry", "Produce", "--input", "input.json", "overflow.gooo"}, reader, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var report struct {
		Decision  string `json:"decision"`
		Reason    string `json:"reason"`
		Failure   struct {
			Code   string `json:"code"`
			Stage  string `json:"stage"`
			Step   string `json:"step"`
			Detail string `json:"detail"`
		} `json:"failure"`
		Execution struct {
			ApplyCalls int            `json:"apply_calls"`
			Activities []string       `json:"activities"`
			Results    map[string]any `json:"results"`
		} `json:"execution"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Reason != "VALUE_INTEGER_OVERFLOW" || report.Failure.Code != report.Reason || report.Failure.Stage != "EXECUTE" || report.Failure.Step != "apply-int-add" || report.Execution.ApplyCalls != 1 || len(report.Execution.Activities) != 1 || report.Execution.Activities[0] != "Produce" || len(report.Execution.Results) != 0 {
		t.Fatalf("partial failure report = %#v", report)
	}
}

type runSourceReaderWithFiles map[string][]byte

func (reader runSourceReaderWithFiles) ReadFile(path string) ([]byte, error) {
	data, ok := reader[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}
