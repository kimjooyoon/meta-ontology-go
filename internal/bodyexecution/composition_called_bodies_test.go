package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCalledBodyConstructionRecordAndObservation(t *testing.T) {
	source, suite := pureCallCompositionFixture(t)
	source = append(source, []byte("\nactivity Main(Integer, Integer, Integer, Text) -> Diagnostic computes `return Diagnose(input0, input1, input2, input3)`\n")...)
	for i := range suite.Cases {
		row := &suite.Cases[i]
		for key, value := range row.Inputs {
			delete(row.Inputs, key)
			row.Inputs[strings.Replace(key, "Diagnose", "Main", 1)] = value
		}
		row.Expected["Main"] = row.Expected["Diagnose"]
		delete(row.Expected, "Diagnose")
	}
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "called.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Preparations) != 1 || len(prior.Steps) != 1 || prior.Plan.Preparations[0].Name != "Diagnose" ||
		prior.Preparations[0].Generation.Report.RecordAssembly.Passed != 5 || prior.Steps[0].Generation.Report.CallClosure == nil {
		t.Fatal("called body was not constructed before its caller", prior.Plan)
	}
	native, err := ExecuteComposition(ctx, "called.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 4 || native.InputSeparation.Status != "UNKNOWN" ||
		native.InputSeparation.Reason != "CALLED_ASSEMBLY_INPUTS_NOT_OBSERVED" || native.InputSeparation.UnknownInputs != 4 {
		t.Fatal("called construction execution or input scope", err, native)
	}
	rows, err := ObserveConstruction(ctx, source, prior)
	if err != nil || len(rows) != 4 || rows[0].Activity != "Diagnose" {
		t.Fatal("helper construction observations disappeared", err, rows)
	}
	raw, _ := json.Marshal(prior)
	decoded, err := DecodeComposition(raw)
	if err != nil || VerifyComposition(ctx, "called.gooo", source, decoded) != nil {
		t.Fatal("called construction did not replay", err)
	}
	decoded.Preparations[0].InputSourceSHA256 = "changed"
	if VerifyComposition(ctx, "called.gooo", source, decoded) == nil {
		t.Fatal("changed helper checkpoint replayed")
	}
}

const calledRecordSeed = `package called
namespace called
entity Integer id "called://integer"
entity Box id "called://box" fields { field value id "called://box/value" type integer required one }
activity Seed(Integer) -> Box computes "return Box{value: input}" assembling {
    choice "value" field_value at "0" alternative "input + 1" intent "Increment."
    value_case "[1]" -> "{\"value\":2}"
    value_case "[2]" -> "{\"value\":3}"
    attempts "2"
}
`

func calledCompositionCases(t *testing.T, raw string) CompositionCases {
	t.Helper()
	suite, err := DecodeCompositionCases([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return suite
}

func TestCalledBodyConstructionOrdersNestedHelpersOnce(t *testing.T) {
	source := []byte(calledRecordSeed + `
activity Wrap(Integer) -> Box computes "let next = Seed(input); return Box{value: next.value}" assembling {
    choice "value" field_value at "0" alternative "next.value * 2" intent "Double the increment."
    value_case "[1]" -> "{\"value\":4}"
    value_case "[2]" -> "{\"value\":6}"
    attempts "2"
}
activity Main(Integer) -> Integer computes "let left = Wrap(input); let right = Wrap(input + 1); return left.value + right.value"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":3},"expected":{"Main":18}}]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "nested.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Preparations) != 2 || prior.Plan.Preparations[0].Name != "Seed" || prior.Plan.Preparations[1].Name != "Wrap" {
		t.Fatal("dependency order or reuse differs", prior.Plan.Preparations)
	}
	native, err := ExecuteComposition(ctx, "nested.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 1 {
		t.Fatal(err, native)
	}
	prior.Preparations[0], prior.Preparations[1] = prior.Preparations[1], prior.Preparations[0]
	if VerifyComposition(ctx, "nested.gooo", source, prior) == nil {
		t.Fatal("changed helper order replayed")
	}
}

func TestCalledBodyConstructionKeepsBoundProducerAndCallArgumentsSeparate(t *testing.T) {
	source := []byte(calledRecordSeed + "activity Main(Box) -> Box computes `return Seed(input.value)`\nbind Seed.result -> Main.input\n")
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Seed":2},"expected":{"Seed":{"value":3},"Main":{"value":4}}}]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "bound.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Preparations) != 1 || len(prior.Steps) != 2 || !prior.Plan.Activities[0].Prepared ||
		prior.Steps[0].Generation.Report.RecordAssembly != nil {
		t.Fatal("bound helper was constructed twice", prior.Plan)
	}
	native, err := ExecuteComposition(ctx, "bound.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 2 {
		t.Fatal(err, native)
	}
	rows, err := ObserveConstruction(ctx, source, prior)
	if err != nil || len(rows) != 2 {
		t.Fatal("construction attempts duplicated", err, rows)
	}
}

func TestCalledBodyConstructionSourceFillAndSearch(t *testing.T) {
	for _, family := range []string{"fill", "search"} {
		t.Run(family, func(t *testing.T) {
			path := "../../examples/body-codegen/source-" + family + "-composition.gooo.fixture"
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			text := string(source)
			text = text[:strings.Index(text, "activity Positive")]
			call, expected := "Scale(Lift(input))", "8"
			if family == "search" {
				call, expected = "AddOffset(Normalize(input))", "5"
			}
			source = []byte(text + "activity Main(Integer) -> Integer computes `return " + call + "`\n")
			suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":3},"expected":{"Main":`+expected+`}}]}`)
			ctx := context.Background()
			prior, err := GenerateCompositionWithOptions(ctx, "called.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
			if err != nil || len(prior.Preparations) != 2 {
				t.Fatal("called profile construction", err, prior.Plan)
			}
			native, err := ExecuteComposition(ctx, "called.gooo", source, prior, suite, nativeTool())
			if err != nil || native.FinitePassed != 1 || native.ModelCalls != 0 {
				t.Fatal("called profile did not replay", err, native)
			}
		})
	}
}

func TestCalledBodyConstructionRejectsCyclesAndMissingPreparation(t *testing.T) {
	source := []byte(calledRecordSeed + "activity Main(Integer) -> Box computes `return Seed(input)`\n")
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":2},"expected":{"Main":{"value":3}}}]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "called.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	stop, _ := resumePolicyFixtures(t)
	if _, err := ResumeComposition(ctx, "called.gooo", source, prior, suite, stop); err == nil || !strings.Contains(err.Error(), "call dependencies") {
		t.Fatal("called construction was silently dropped during continuation", err)
	}
	prior.Preparations = nil
	if VerifyComposition(ctx, "called.gooo", source, prior) == nil {
		t.Fatal("omitted helper construction replayed")
	}
	cycle := []byte(strings.Replace(string(source), `return Box{value: input}`, `return Main(input)`, 1))
	failed, err := GenerateCompositionWithOptions(ctx, "cycle.gooo", cycle, suite,
		CompositionOptions{EntryActivity: "Main", ModelPath: "/missing-model"})
	if err == nil || !strings.Contains(err.Error(), "recursive") || failed.Model != nil || len(failed.Preparations) != 0 {
		t.Fatal("recursive assembly reached inference", err)
	}
}
