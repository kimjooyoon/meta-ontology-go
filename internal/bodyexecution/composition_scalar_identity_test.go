package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCompositionScalarIdentitiesSurviveTypeRenames(t *testing.T) {
	for _, test := range []struct{ before, after, id, body, input, expected string }{
		{"Integer", "정수", "integer", "return input + 7", "9007199254740993", "9007199254741000"},
		{"Boolean", "논리", "boolean", "return !input", "true", "false"},
		{"Text", "문자열", "string", `return input + "!"`, `"한글"`, `"한글!"`},
	} {
		var generated string
		for _, name := range []string{test.before, test.after} {
			source := []byte(fmt.Sprintf("package p\nnamespace p\nentity %s id %q\nactivity Apply(%s) -> %s id \"urn:gooo:apply\" computes %q\n", name, "urn:gooo:type:"+test.id, name, name, test.body))
			suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{
				Inputs: map[string]json.RawMessage{"Apply": json.RawMessage(test.input)}, Expected: map[string]json.RawMessage{"Apply": json.RawMessage(test.expected)},
			}}}
			prior, err := GenerateCompositionWithOptions(context.Background(), "alias.gooo", source, suite, CompositionOptions{EntryActivity: "Apply"})
			if err != nil {
				t.Fatal(name, err)
			}
			node := prior.Plan.Activities[0]
			if node.InputType != name || node.OutputType != name || node.InputEntityID != "urn:gooo:type:"+test.id || !strings.Contains(prior.GoooSource, "entity "+name) {
				t.Fatal("authored names or scalar identities changed", name, node)
			}
			if generated != "" && generated != prior.GeneratedSHA256 {
				t.Fatal("renaming the scalar changed native code", name)
			}
			generated = prior.GeneratedSHA256
			run, err := ExecuteComposition(context.Background(), "alias.gooo", source, prior, suite, nativeTool())
			if err != nil || run.FinitePassed != 1 || run.FiniteTotal != 1 || !run.ProjectionReplayed || !run.RuntimeReplayed || run.ModelCalls != 0 {
				t.Fatal("scalar alias native execution or saved replay differed", name, err, run)
			}
			if string(run.Traces[0].Deliveries[0].Actual) != test.expected || run.Traces[0].Deliveries[0].ActivityID != "urn:gooo:apply" {
				t.Fatal("native delivery lost exact value or activity identity", name)
			}
		}
	}
}
