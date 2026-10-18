package workspaceexecution

import (
	"context"
	"encoding/json"
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
		name, input, want string
	}{
		{"missing", `{}`, `root input "example/core:Normalize" requires an explicit value`},
		{"wrong port", `{"example/core:Normalize.input0":0}`, `use "example/core:Normalize"`},
		{"bound port", `{"example/core:Normalize":"hello","example/app:Main":"hello"}`, `bound input "example/app:Main"`},
		{"wrong type", `{"example/core:Normalize":true}`, `input "example/core:Normalize"`},
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
		})
	}
	trailing := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{
			"example/core:Normalize.": json.RawMessage(`"hello"`),
		}}}}
	if _, err := translateCases(program, trailing); err == nil || !strings.Contains(err.Error(), `example/core:Normalize.`) {
		t.Fatal("empty port suffix was accepted as a valid single-input key", err)
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
	valid := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{
			"example/math:Main.input0": json.RawMessage(`1`),
			"example/math:Main.input1": json.RawMessage(`2`),
		}}}}
	translated, err := translateCases(program, valid)
	if err != nil {
		t.Fatal(err)
	}
	if err := bodyexecution.ValidateCompositionCases(context.Background(), "workspace.gooo", []byte(program.Source), translated); err != nil {
		t.Fatal("diagnostic lens changed accepted input", err)
	}
}
