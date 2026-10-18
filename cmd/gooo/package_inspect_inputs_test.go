package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPackageInspectInputTemplateUsesAuthoredKeys(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	module := "example.org/diagnostic"
	reader := mapSourceReader{}
	for _, name := range []string{"gooo.workspace.json", "app.gooo.template", "diagnostics.gooo.template"} {
		data := string(read("templates/diagnostic/" + name))
		data = strings.ReplaceAll(data, "{{module}}", module)
		data = strings.ReplaceAll(data, "{{namespace}}", "sample")
		key := "sample/" + name
		if strings.HasSuffix(key, ".template") {
			key = strings.TrimSuffix(key, ".template")
		}
		reader[key] = []byte(data)
	}
	var out, diagnostic bytes.Buffer
	args := []string{"--inputs-template", "sample/gooo.workspace.json"}
	if code := runPackageInspect(args, reader, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String())
	}
	var template struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}
	if err := json.Unmarshal(out.Bytes(), &template); err != nil {
		t.Fatal(err)
	}
	if template.Schema != "gooo/body-composition-inputs/v1" || len(template.Inputs) != 1 || len(template.Inputs[0]) != 4 {
		t.Fatal(template.Schema, template.Inputs)
	}
	for key, want := range map[string]string{
		module + "/diagnostics:Diagnose.input0": "0",
		module + "/diagnostics:Diagnose.input1": "0",
		module + "/diagnostics:Diagnose.input2": "0",
		module + "/diagnostics:Diagnose.input3": `""`,
	} {
		if got := string(template.Inputs[0][key]); got != want {
			t.Errorf("%s = %s; want %s", key, got, want)
		}
	}
	args = []string{"--json", "--inputs-template", "sample/gooo.workspace.json"}
	if code := runPackageInspect(args, reader, &out, &diagnostic); code != exitUsage {
		t.Fatalf("mixed output modes = %d", code)
	}
}

func TestPackageInspectInputTemplateKeepsSingleInputKey(t *testing.T) {
	const example = "../../examples/workspace-call-identity/"
	reader := mapSourceReader{
		"sample/workspace.json":    mustReadPackageInspectFixture(t, example+"workspace.json"),
		"sample/app.gooo.fixture":  mustReadPackageInspectFixture(t, example+"app.gooo.fixture"),
		"sample/core.gooo.fixture": mustReadPackageInspectFixture(t, example+"core.gooo.fixture"),
	}
	var out, diagnostic bytes.Buffer
	args := []string{"sample/workspace.json", "--inputs-template"}
	if code := runPackageInspect(args, reader, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String())
	}
	var template struct {
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}
	if err := json.Unmarshal(out.Bytes(), &template); err != nil || len(template.Inputs) != 1 {
		t.Fatal(err, out.String())
	}
	if got := string(template.Inputs[0]["example/app:Main"]); got != "0" || len(template.Inputs[0]) != 1 {
		t.Fatal(template.Inputs[0])
	}
}

func mustReadPackageInspectFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
