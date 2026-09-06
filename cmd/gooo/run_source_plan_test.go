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

type runSourceReaderWithFiles map[string][]byte

func (reader runSourceReaderWithFiles) ReadFile(path string) ([]byte, error) {
	data, ok := reader[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}
