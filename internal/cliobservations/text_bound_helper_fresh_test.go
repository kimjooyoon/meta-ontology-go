package cliobservations

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func textBoundWorkspace1905(t *testing.T, source []byte, activity string) string {
	t.Helper()
	dir := t.TempDir()
	manifest := map[string]any{"schema": "gooo/package-workspace-manifest/v1",
		"entry":    map[string]string{"package_path": "app/bound1905", "activity": activity},
		"packages": []any{map[string]any{"path": "app/bound1905", "sources": []string{"main.gooo"}}}}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string][]byte{"workspace.json": raw, "main.gooo": source} {
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "workspace.json")
}

func executeTextBound1905(t *testing.T, name string, source []byte, activity string,
	suite bodyexecution.CompositionCases, inputOnly bool) (packageExecutionReceipt, int) {
	t.Helper()
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	manifest := textBoundWorkspace1905(t, source, activity)
	raw, err := json.Marshal(suite)
	if err != nil {
		t.Fatal(err)
	}
	flag := "--cases"
	if inputOnly {
		flag = "--inputs"
		rows := make([]map[string]json.RawMessage, len(suite.Cases))
		for i, row := range suite.Cases {
			rows[i] = row.Inputs
		}
		raw, err = json.Marshal(map[string]any{"schema": bodyexecution.CompositionInputsSchema, "inputs": rows})
		if err != nil {
			t.Fatal(err)
		}
	}
	input := filepath.Join(filepath.Dir(manifest), "cases.json")
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var output, diagnostic bytes.Buffer
	code := run([]string{"package", "execute", "--json", flag, input, manifest}, &output, &diagnostic)
	retainTextBound1905(t, name, map[string]any{"source": string(source), "cases": json.RawMessage(raw),
		"exit_code": code, "stdout": output.String(), "stderr": diagnostic.String()})
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil {
		t.Fatal("CLI result was not JSON", code, err, output.String(), diagnostic.String())
	}
	return receipt, code
}

func retainTextBound1905(t *testing.T, name string, observation any) {
	t.Helper()
	root := os.Getenv("GOOO_CI_ORIGIN_ROOT")
	if root == "" {
		return
	}
	dir := filepath.Join(root, "text-bound-native")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	metadata := map[string]any{"source_head": os.Getenv("GOOO_CI_SOURCE_HEAD"),
		"run_id": os.Getenv("GITHUB_RUN_ID"), "attempt": os.Getenv("GITHUB_RUN_ATTEMPT"),
		"job": os.Getenv("GITHUB_JOB"), "observation": observation}
	metadata["source_head_scope"] = "ADMITTED_WORK"
	metadata["compiler_build_source_sha"] = commandSourceIdentity
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}
