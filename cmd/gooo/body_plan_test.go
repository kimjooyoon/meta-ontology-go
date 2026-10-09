package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBodyPlanShowsInputsWithoutGenerating(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	source := "../../examples/scalar-identity/source.gooo.fixture"
	for _, mode := range []string{"", "--json", "--inputs-template"} {
		args := []string{"body-plan", "--source", source, "--entry", "Describe"}
		if mode != "" {
			args = append(args, mode)
		}
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String())
		}
		switch mode {
		case "":
			for _, want := range []string{"Describe.input0", "Describe.input1", "Describe.input2", "record_choices", "0 model calls"} {
				if !strings.Contains(out.String(), want) {
					t.Fatal("missing readable plan", want, out.String())
				}
			}
		case "--json":
			var result bodyexecution.CompositionInspection
			if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Schema != "gooo/body-composition-inspection/v1" ||
				result.CandidateTests != 0 || len(result.CallerInputs) != 3 {
				t.Fatal(err, out.String())
			}
		case "--inputs-template":
			if _, err := bodyexecution.DecodeCompositionInputs(out.Bytes()); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "expected") {
				t.Fatal("template acquired expected outputs")
			}
		}
	}
}

func TestBodyPlanRejectsGenerationFlagsAndAmbiguousOutput(t *testing.T) {
	for _, suffix := range [][]string{{"--json", "--inputs-template"}, {"--model", "never-load-model.json"},
		{"--cases", "never-load-cases.json"}, {"--entry", "Describe", "--entry", "Describe"}, {"--out", "never-create"}} {
		var out, diagnostics bytes.Buffer
		args := append([]string{"body-plan", "--source", "missing.gooo"}, suffix...)
		if code := run(args, &out, &diagnostics); code != exitUsage {
			t.Fatal("invalid flags reached source", code, diagnostics.String())
		}
	}
}
