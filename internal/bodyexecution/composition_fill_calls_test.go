package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCompositionSourceFillCallsFixedHelper(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-calls/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":10}}]`)
	prior, err := GenerateCompositionWithOptions(ctx, "fill.gooo", source, cases, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal("fill helper was not available during composition", err)
	}
	if len(prior.Preparations) != 1 || prior.Preparations[0].Generation.Report.BodyFill == nil ||
		prior.Preparations[0].Generation.Report.BodyFill.SelectedCandidateID != "shifted" {
		t.Fatal("composition lost its selected fill helper", prior.Preparations)
	}
	raw, _ := json.MarshalIndent(prior, "", "  ")
	saved, err := DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(ctx, "fill.gooo", source, saved, cases, nativeTool())
	if err != nil || runtime.FinitePassed != 1 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("source-filled helper did not reach native execution", runtime, err)
	}
}

func TestCompositionFillWaitsForTransitiveAssembledHelper(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-calls/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	prepared := strings.Replace(string(source), `activity Step(Integer) -> Integer computes "return input * 2"`, `
entity Box id "fillcalls://box" fields { field value id "fillcalls://value" type integer required one }
activity Seed(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "input * 2" intent "Double the seed."
 value_case "[7]" -> "{\"value\":14}"
 attempts "2"
}
activity Step(Integer) -> Integer computes "let seed = Seed(input); return seed.value"
`, 1)
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":10}}]`)
	prior, err := GenerateCompositionWithOptions(ctx, "prepared.gooo", []byte(prepared), cases,
		CompositionOptions{EntryActivity: "Main"})
	if err != nil || len(prior.Preparations) != 2 || prior.Preparations[0].Generation.Report.Activity != "Seed" ||
		prior.Preparations[1].Generation.Report.BodyFill == nil {
		t.Fatal("transitive seed was not realized before body fill", prior.Preparations, err)
	}
	runtime, err := ExecuteComposition(ctx, "prepared.gooo", []byte(prepared), prior, cases, nativeTool())
	if err != nil || runtime.FinitePassed != 1 || runtime.ModelCalls != 0 || runtime.InputSeparation.Status != "UNKNOWN" {
		t.Fatal("prepared helper or unobserved fill exposure differs", runtime, err)
	}
}
