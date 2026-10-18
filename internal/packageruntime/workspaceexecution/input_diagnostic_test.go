package workspaceexecution

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceInputDiagnosticsUseAuthoredKeys(t *testing.T) {
	program, err := Prepare(importedActivityWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		name, input, want, code, key, declared string
	}{
		{"missing", `{}`, `root input "example/core:Normalize" requires an explicit value`, "MISSING_ROOT_INPUT", "", "example/core:Normalize"},
		{"wrong port", `{"example/core:Normalize.input0":0}`, `use "example/core:Normalize"`, "UNKNOWN_CALLER_PORT", "example/core:Normalize.input0", "example/core:Normalize"},
		{"bound port", `{"example/core:Normalize":"hello","example/app:Main":"hello"}`, `bound input "example/app:Main"`, "BOUND_INPUT_PROVIDED", "example/app:Main", ""},
		{"wrong type", `{"example/core:Normalize":true}`, `input "example/core:Normalize"`, "INVALID_INPUT_VALUE", "example/core:Normalize", ""},
	} {
		t.Run(check.name, func(t *testing.T) {
			var inputs map[string]json.RawMessage
			if err := json.Unmarshal([]byte(check.input), &inputs); err != nil {
				t.Fatal(err)
			}
			suite := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
				Cases: []bodyexecution.CompositionCase{{Inputs: inputs}}}
			translated, err := translateCases(program, suite)
			if err != nil {
				t.Fatal(err)
			}
			err = bodyexecution.ValidateCompositionCases(context.Background(), "workspace.gooo", []byte(program.Source), translated)
			if err == nil || !strings.Contains(err.Error(), check.want) || strings.Contains(err.Error(), "GoooPackage") {
				t.Fatalf("diagnostic = %v; want %q without lowered name", err, check.want)
			}
			var diagnostic *bodyexecution.InputDiagnostic
			if !errors.As(err, &diagnostic) || diagnostic.Stage != "INPUT_VALIDATION" ||
				diagnostic.Code != check.code || diagnostic.CaseIndex != 0 ||
				diagnostic.Key != check.key || diagnostic.DeclaredKey != check.declared {
				t.Fatalf("structured diagnostic = %+v, error = %v", diagnostic, err)
			}
		})
	}
	trailing := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{
			"example/core:Normalize.": json.RawMessage(`"hello"`),
		}}}}
	if _, err := translateCases(program, trailing); err == nil || !strings.Contains(err.Error(), `example/core:Normalize.`) {
		t.Fatal("empty port suffix was accepted as a valid single-input key", err)
	}
	unknown := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{
			"z:Unknown": json.RawMessage(`1`), "a:Unknown": json.RawMessage(`1`),
		}}}}
	for range 8 {
		_, err := translateCases(program, unknown)
		var diagnostic *bodyexecution.InputDiagnostic
		if !errors.As(err, &diagnostic) || diagnostic.Stage != "INPUT_RESOLUTION" ||
			diagnostic.Code != "UNKNOWN_ACTIVITY_KEY" || diagnostic.Key != "a:Unknown" {
			t.Fatalf("first unknown key is not stable: %+v, %v", diagnostic, err)
		}
	}
	expectedPort := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1",
		Cases: []bodyexecution.CompositionCase{{Expected: map[string]json.RawMessage{
			"example/app:Main.input0": json.RawMessage(`1`),
		}}}}
	_, err = translateCases(program, expectedPort)
	var diagnostic *bodyexecution.InputDiagnostic
	if !errors.As(err, &diagnostic) || diagnostic.Code != "EXPECTED_PORT_NOT_ALLOWED" ||
		diagnostic.Key != "example/app:Main.input0" {
		t.Fatalf("expected port diagnostic = %+v, %v", diagnostic, err)
	}
}

func TestWorkspaceExpectedDiagnosticUsesAuthoredKey(t *testing.T) {
	program, err := Prepare(importedActivityWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1",
		Cases: []bodyexecution.CompositionCase{{
			Inputs:   map[string]json.RawMessage{"example/core:Normalize": json.RawMessage(`"hello"`)},
			Expected: map[string]json.RawMessage{"example/app:Main": json.RawMessage(`true`)},
		}}}
	translated, err := translateCases(program, suite)
	if err != nil {
		t.Fatal(err)
	}
	err = bodyexecution.ValidateCompositionCases(context.Background(), "workspace.gooo", []byte(program.Source), translated)
	if err == nil || !strings.Contains(err.Error(), `expected "example/app:Main"`) ||
		strings.Contains(err.Error(), "GoooPackage") {
		t.Fatal("expected-value diagnostic lost its authored key", err)
	}
	var diagnostic *bodyexecution.InputDiagnostic
	if !errors.As(err, &diagnostic) || diagnostic.Code != "INVALID_EXPECTED_VALUE" ||
		diagnostic.Key != "example/app:Main" {
		t.Fatal("expected-value diagnostic lost its structured source key", diagnostic, err)
	}
}

func TestWorkspaceInputDiagnosticsKeepCaseJSONAndMultiportShape(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/math", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{{Path: "example/math", Name: "math", Sources: []packageruntime.Source{{
			Filename: "main.gooo", Content: `package math
namespace math
entity Integer id "urn:gooo:type:integer"
activity Main(Integer,Integer) -> Integer computes "return input0 + input1"
`,
		}}}},
	}
	program, err := Prepare(manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		inputs, want string
	}{
		{`{"example/math:Main.input0":1}`, `root input "example/math:Main.input1"`},
		{`{"example/math:Main.input0":1,"example/math:Main.input2":2}`, `available ports: ["example/math:Main.input0" "example/math:Main.input1"]`},
	} {
		var inputs map[string]json.RawMessage
		if err := json.Unmarshal([]byte(check.inputs), &inputs); err != nil {
			t.Fatal(err)
		}
		suite := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
			Cases: []bodyexecution.CompositionCase{{Inputs: inputs}}}
		translated, err := translateCases(program, suite)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(translated)
		if err != nil || strings.Contains(string(raw), "Display") || strings.Contains(string(raw), "example/math:Main") {
			t.Fatal("diagnostic lens leaked into executable cases", string(raw), err)
		}
		err = bodyexecution.ValidateCompositionCases(context.Background(), "workspace.gooo", []byte(program.Source), translated)
		if err == nil || !strings.Contains(err.Error(), check.want) || strings.Contains(err.Error(), "GoooPackage") {
			t.Fatalf("diagnostic = %v; want %q without lowered name", err, check.want)
		}
	}
	multiCase := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{
			{Inputs: map[string]json.RawMessage{"example/math:Main.input0": json.RawMessage(`1`),
				"example/math:Main.input1": json.RawMessage(`2`)}},
			{Inputs: map[string]json.RawMessage{"example/math:Main.input2": json.RawMessage(`3`)}},
		}}
	translated, err := translateCases(program, multiCase)
	if err != nil {
		t.Fatal(err)
	}
	err = bodyexecution.ValidateCompositionCases(context.Background(), "workspace.gooo", []byte(program.Source), translated)
	var diagnostic *bodyexecution.InputDiagnostic
	if !errors.As(err, &diagnostic) || diagnostic.Code != "UNKNOWN_CALLER_PORT" ||
		diagnostic.CaseIndex != 1 || diagnostic.Key != "example/math:Main.input2" ||
		len(diagnostic.AvailableKeys) != 2 || diagnostic.AvailableKeys[0] != "example/math:Main.input0" ||
		diagnostic.AvailableKeys[1] != "example/math:Main.input1" {
		t.Fatalf("second case caller ports = %+v, %v", diagnostic, err)
	}
	valid := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{
			"example/math:Main.input0": json.RawMessage(`1`),
			"example/math:Main.input1": json.RawMessage(`2`),
		}}}}
	translated, err = translateCases(program, valid)
	if err != nil {
		t.Fatal(err)
	}
	if err := bodyexecution.ValidateCompositionCases(context.Background(), "workspace.gooo", []byte(program.Source), translated); err != nil {
		t.Fatal("diagnostic lens changed accepted input", err)
	}
}
