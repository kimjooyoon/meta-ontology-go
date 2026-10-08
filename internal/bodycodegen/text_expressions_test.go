package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const textExpressionPrelude = `package textops
namespace textops
entity Text id "gooo://textops/text"
entity Integer id "gooo://textops/integer"
entity Boolean id "gooo://textops/boolean"
entity Result id "gooo://textops/result" fields {
    field text id "gooo://textops/result/text" type string required one
    field size id "gooo://textops/result/size" type integer required one
    field matches id "gooo://textops/result/matches" type boolean required one
}
`

func TestTextOperationsRetainSymbolicInputOrigins(t *testing.T) {
	source := []byte(textExpressionPrelude + "activity Select(Text) -> Result computes `return Result{text: input[:], size: int64(len(input)), matches: true}` assembling {\n" +
		"choice \"text\" field_value at \"0\" alternative \"input[:0]\" intent \"Keep the entire input.\"\n" +
		"choice \"size\" field_value at \"1\" alternative \"0\" intent \"Keep the byte length.\"\n" +
		"choice \"matches\" field_value at \"2\" alternative \"false\" intent \"Keep true.\"\n" +
		"value_case \"[\\\"가\\\"]\" -> \"{\\\"text\\\":\\\"가\\\",\\\"size\\\":3,\\\"matches\\\":true}\"\nattempts \"8\"\n}\n")
	result, err := ExportRecordAssemblyContextWithFlow(context.Background(), "origins.gooo", source, "Select", false, jointdecision.RecordSharedFeatureVersion)
	if err != nil || result.ValueFlow == nil || result.ValueFlow.Status != "RESOLVED" {
		t.Fatal("text origin graph", result.ValueFlow, err)
	}
	flow := result.ValueFlow
	if !flowContains(flow, flow.Choices[0].First, "expression", "slice_bounds") ||
		!flowContains(flow, flow.Choices[0].First, "expression", "len") ||
		!flowContains(flow, flow.Choices[1].First, "expression", "int64") ||
		!flowContains(flow, flow.Choices[1].First, "input", "") || result.CandidateTests != 0 || result.ModelPredictions != 0 {
		t.Fatal("text origins lost offsets, conversion, or source-only observation", result)
	}
}

func TestTextOperationsGenerateInterpretAndReplay(t *testing.T) {
	source := []byte(textExpressionPrelude + "activity Prefix(Text, Text) -> Boolean computes `return len(input0) >= len(input1) && input0[:len(input1)] == input1`\n" +
		"activity Run(Text, Text) -> Result computes `let n = len(input0); let size = int64(n); return Result{text: input0[:], size: size, matches: Prefix(input0, input1)}`\n")
	generated, err := Generate("text.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	closure := generated.Report.CallClosure
	if !generated.Report.RouteEquivalence.Equivalent || !generated.Report.DeterministicReplay ||
		closure == nil || len(closure.Activities) != 1 || len(closure.Edges) != 1 || closure.MaxCallsPerInvocation != 1 {
		t.Fatal("text primitives changed source-call identity", generated.Report)
	}
	cases := []assemblyspec.ValueCase{
		{Inputs: `["gooo", "go"]`, Expected: `{"text":"gooo","size":4,"matches":true}`},
		{Inputs: `["가나다", "가"]`, Expected: `{"text":"가나다","size":9,"matches":true}`},
		{Inputs: `["", ""]`, Expected: `{"text":"","size":0,"matches":true}`},
		{Inputs: `["g", "gooo"]`, Expected: `{"text":"g","size":1,"matches":false}`},
		{Inputs: `["🙂gooo", "🙂"]`, Expected: `{"text":"🙂gooo","size":8,"matches":true}`},
		{Inputs: `["é", "é"]`, Expected: `{"text":"é","size":3,"matches":false}`},
	}
	results, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes, cases)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if !result.Passed {
			t.Fatal("text observation differs", result)
		}
	}
}

func TestTextSliceBoundsAndEvaluation(t *testing.T) {
	for _, test := range []struct{ expression, inputs, expected string }{
		{"input0[input1:input2]", `["가나다",3,6]`, "나"},
		{"input0[:input2]", `["hello",0,2]`, "he"},
		{"input0[input1:]", `["hello",2,5]`, "llo"},
		{"input0[:]", `["hello",0,5]`, "hello"},
		{"input0[input1:input1]", `["hello",5,5]`, ""},
	} {
		t.Run(test.expression, func(t *testing.T) {
			body := "return Result{text: " + test.expression + ", size: int64(len(input0)), matches: true}"
			source := []byte(textExpressionPrelude + "activity Run(Text, Integer, Integer) -> Result computes `" + body + "`\n")
			generated, err := Generate("slice.gooo", source, "Run")
			if err != nil {
				t.Fatal(err)
			}
			size := 5
			if strings.Contains(test.inputs, "가나다") {
				size = 9
			}
			cases := []assemblyspec.ValueCase{{Inputs: test.inputs, Expected: fmt.Sprintf(`{"text":%q,"size":%d,"matches":true}`, test.expected, size)}}
			results, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes, cases)
			if err != nil || !results[0].Passed {
				t.Fatal(results, err)
			}
		})
	}
	source := []byte(textExpressionPrelude + "activity Run(Text, Integer, Integer) -> Result computes `return Result{text: input0[input1:input2], size: 0, matches: true}`\n")
	generated, err := Generate("bounds.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	for _, inputs := range []string{`["gooo",-1,2]`, `["gooo",3,2]`, `["gooo",0,5]`, `["gooo",0,9007199254740993]`} {
		_, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes,
			[]assemblyspec.ValueCase{{Inputs: inputs, Expected: `{"text":"","size":0,"matches":true}`}})
		if err == nil || !strings.Contains(err.Error(), "Text slice byte offsets") {
			t.Fatal("invalid offsets did not fail evaluation", inputs, err)
		}
	}
}

func TestTextOperationsKeepTypeAndCallBoundaries(t *testing.T) {
	for _, body := range []string{
		`return input[0:1:2]`, `return input[-1:]`, `return input[false:]`,
		`return len(input)`, `return int64(input)`, `return input[0]`,
		`return len(input, input)`, `return input[:len(true)]`,
	} {
		source := []byte(textExpressionPrelude + "activity Run(Text) -> Text computes `" + body + "`\n")
		if _, err := Generate("invalid.gooo", source, "Run"); err == nil {
			t.Fatal("unsupported expression accepted", body)
		}
	}
	source := []byte(textExpressionPrelude + "activity len(Text) -> Integer computes `return 42`\n" +
		"activity Run(Text) -> Result computes `return Result{text: input, size: len(input), matches: true}`\n")
	generated, err := Generate("shadow.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	if generated.Report.CallClosure == nil || generated.Report.CallClosure.Activities[0].Name != "len" {
		t.Fatal("source-declared len lost its identity")
	}
	results, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes,
		[]assemblyspec.ValueCase{{Inputs: `["gooo"]`, Expected: `{"text":"gooo","size":42,"matches":true}`}})
	if err != nil || !results[0].Passed {
		t.Fatal("source-declared len was treated as a primitive", results, err)
	}
}
