package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCompositionInputOnlyPreservesDisjointInputsWithoutClaimingPasses(t *testing.T) {
	source, cases := compositionFixture(t)
	prior, err := GenerateComposition(context.Background(), "observe.gooo", source, cases, "")
	if err != nil {
		t.Fatal(err)
	}
	input := CompositionCases{Schema: CompositionInputsSchema, Cases: cases.Cases}
	for i := range input.Cases {
		input.Cases[i].Expected = nil
	}
	r, err := ExecuteComposition(context.Background(), "observe.gooo", source, prior, input, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	if r.FiniteTotal != 0 || r.FinitePassed != 0 || !r.RuntimeReplayed || r.InputSeparation.Status != "UNKNOWN" ||
		r.InputSeparation.Reason != "NO_RUNTIME_EXPECTATIONS" || r.InputSeparation.DisjointCasesPassed != 0 || r.InputSeparation.UniqueInputs == 0 {
		t.Fatal("unverified outputs were counted as passing cases", r)
	}
	input.Cases[0].Expected = map[string]json.RawMessage{"Assemble": json.RawMessage(`0`)}
	if _, err := ExecuteComposition(context.Background(), "observe.gooo", source, prior, input, nativeTool()); err == nil {
		t.Fatal("input-only API accepted an expectation")
	}
}

func TestCompositionInputOnlyDecodeKeepsCaseContractStrict(t *testing.T) {
	valid := `{"schema":"gooo/body-composition-inputs/v1","inputs":[{"Main":9007199254740993}]}`
	suite, err := DecodeCompositionInputs([]byte(valid))
	if err != nil || string(suite.Cases[0].Inputs["Main"]) != "9007199254740993" {
		t.Fatal("exact integer input lost", err)
	}
	if _, err := DecodeCompositionCases([]byte(valid)); err == nil {
		t.Fatal("case-only decoder accepted input-only mode")
	}
	for _, bad := range []string{
		strings.Replace(valid, "inputs/v1", "cases/v1", 1),
		`{"schema":"gooo/body-composition-inputs/v1","inputs":[]}`,
		`{"schema":"gooo/body-composition-inputs/v1","inputs":[{}],"expected":{}}`,
		valid + ` {}`,
	} {
		if _, err := DecodeCompositionInputs([]byte(bad)); err == nil {
			t.Fatal("invalid input document accepted", bad)
		}
	}
	source := []byte("package tool\nnamespace tool\nentity Integer id \"tool://integer\"\nactivity Main(Integer) -> Integer computes \"return input\"\n")
	suite.Schema = "gooo/body-composition-cases/v1"
	if err := ValidateCompositionCases(context.Background(), "tool.gooo", source, suite); err == nil {
		t.Fatal("case mode accepted missing expectations")
	}
	suite.Schema = CompositionInputsSchema
	for _, row := range []map[string]json.RawMessage{
		{}, {"Unknown": json.RawMessage(`1`)}, {"Main": json.RawMessage(`"1"`)}, {"Main": json.RawMessage(`null`)},
	} {
		suite.Cases[0].Inputs = row
		if err := ValidateCompositionCases(context.Background(), "tool.gooo", source, suite); err == nil {
			t.Fatal("input-only mode bypassed typed input validation", row)
		}
	}
}
