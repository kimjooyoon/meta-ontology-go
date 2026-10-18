package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPackageExecuteInputDiagnosticUsesAuthoredKey(t *testing.T) {
	const example = "../../examples/workspace-call-identity/"
	reader := mapSourceReader{
		"sample/workspace.json":    mustReadPackageInspectFixture(t, example+"workspace.json"),
		"sample/app.gooo.fixture":  mustReadPackageInspectFixture(t, example+"app.gooo.fixture"),
		"sample/core.gooo.fixture": mustReadPackageInspectFixture(t, example+"core.gooo.fixture"),
		"sample/inputs.json": []byte(`{"schema":"gooo/body-composition-inputs/v1",` +
			`"inputs":[{"example/app:Main.input0":0}]}`),
	}
	var out, diagnostic bytes.Buffer
	code := runPackageExecute([]string{"--json", "--inputs", "sample/inputs.json", "sample/workspace.json"},
		reader, &out, &diagnostic)
	if code != exitFailure {
		t.Fatal(code, out.String(), diagnostic.String())
	}
	var receipt struct {
		Decision        string `json:"decision"`
		Error           string `json:"error"`
		InputDiagnostic struct {
			Stage       string `json:"stage"`
			Code        string `json:"code"`
			CaseIndex   int    `json:"case_index"`
			Key         string `json:"key"`
			DeclaredKey string `json:"declared_key"`
		} `json:"input_diagnostic"`
	}
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err, out.String())
	}
	if receipt.Decision != "FAIL_CLOSED" ||
		!strings.Contains(receipt.Error, `input "example/app:Main.input0"`) ||
		!strings.Contains(receipt.Error, `use "example/app:Main"`) ||
		strings.Contains(receipt.Error, "GoooPackage") ||
		receipt.InputDiagnostic.Stage != "INPUT_VALIDATION" ||
		receipt.InputDiagnostic.Code != "UNKNOWN_CALLER_PORT" ||
		receipt.InputDiagnostic.CaseIndex != 0 ||
		receipt.InputDiagnostic.Key != "example/app:Main.input0" ||
		receipt.InputDiagnostic.DeclaredKey != "example/app:Main" {
		t.Fatal("input diagnostic lost its authored key", receipt)
	}
}
