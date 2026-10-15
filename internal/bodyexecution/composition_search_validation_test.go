package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

const searchValidationSource = `package validation
namespace validation
entity Integer id "validation://integer"
activity Step(Integer) -> Integer computes "return input * 2"
activity Build(Integer) -> Integer computes "return Step(input + __GOOO_BODY_HOLE_value__)" assembling {
 search hole "value" grammar "integer-hole-residual/v1" intent "Choose an offset." max_candidates "8"
 case "2" -> "6" attempts "8"
}
activity Main(Integer) -> Integer computes "return Build(input)"
`

func TestCompositionRetainsAndReplaysSearchValidationBoundaries(t *testing.T) {
	source := []byte(searchValidationSource)
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":10}}]`)
	prior, err := GenerateCompositionWithOptions(ctx, "validation.gooo", source, cases, CompositionOptions{EntryActivity: "Main"})
	if err != nil || len(prior.SearchValidations) != 2 {
		t.Fatal("original preflight boundaries were discarded", prior.SearchValidations, err)
	}
	for i, stage := range []string{"BODY_PREFLIGHT", "ACTIVITY_GENERATION"} {
		check := prior.SearchValidations[i]
		if check.Stage != stage || !check.Observation.Completed || check.Observation.SourceSHA256 != digest(source) ||
			len(check.Observation.LocalCalls.Evaluations) != 2 {
			t.Fatal("validation checkpoint or actual probes differ", check)
		}
	}
	raw, _ := json.MarshalIndent(prior, "", "  ")
	saved, err := DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyComposition(ctx, "validation.gooo", source, saved); err != nil {
		t.Fatal("saved validation history did not replay", err)
	}
	runtime, err := ExecuteComposition(ctx, "validation.gooo", source, saved, cases, nativeTool())
	if err != nil || runtime.FinitePassed != 1 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("validation history did not reach native execution", runtime, err)
	}
	for name, mutate := range map[string]func(*Composition){
		"stage":  func(c *Composition) { c.SearchValidations[0].Stage = "ACTIVITY_GENERATION" },
		"source": func(c *Composition) { c.SearchValidations[1].Observation.SourceSHA256 = "changed" },
		"argument": func(c *Composition) {
			c.SearchValidations[0].Observation.LocalCalls.Evaluations[0].Calls[0].Inputs = json.RawMessage("[99]")
		},
		"missing check": func(c *Composition) { c.SearchValidations = c.SearchValidations[:1] },
		"extra check":   func(c *Composition) { c.SearchValidations = append(c.SearchValidations, c.SearchValidations[0]) },
	} {
		t.Run(name, func(t *testing.T) {
			changed, err := DecodeComposition(raw)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&changed)
			if err := VerifyComposition(ctx, "validation.gooo", source, changed); err == nil {
				t.Fatal("changed validation evidence was accepted")
			}
		})
	}
	saved.SearchValidations = nil
	if err := VerifyComposition(ctx, "validation.gooo", source, saved); err != nil || saved.SearchValidations != nil {
		t.Fatal("historical absence was rewritten", err)
	}
}

func TestCompositionSearchValidationUsesPreparedHelperCheckpoint(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-search-calls/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	construction := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":4}}]`)
	prior, err := GenerateCompositionWithOptions(context.Background(), "prepared.gooo", source, construction,
		CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.SearchValidations) != 1 || prior.SearchValidations[0].Stage != "ACTIVITY_GENERATION" ||
		prior.SearchValidations[0].Observation.Activity != "Divide" ||
		prior.SearchValidations[0].Observation.SourceSHA256 == digest(source) ||
		prior.SearchValidations[0].Observation.SourceSHA256 != prior.Preparations[1].InputSourceSHA256 {
		t.Fatal("deferred validation did not bind the realized helper checkpoint", prior.SearchValidations)
	}
	if err := VerifyComposition(context.Background(), "prepared.gooo", source, prior); err != nil {
		t.Fatal("prepared source validation could not be replayed", err)
	}
}

func TestCompositionSearchValidationKeepsEarlierPreflightBeforeFailure(t *testing.T) {
	source := []byte(searchValidationSource + `
activity Broken(Integer) -> Integer computes "return missing"
bind Build.result -> Broken.input
`)
	cases := jointCases(t, `[{"inputs":{"Build":4},"expected":{"Broken":10}}]`)
	prior, err := GenerateCompositionWithOptions(context.Background(), "partial.gooo", source, cases,
		CompositionOptions{EntryActivity: "Broken"})
	if err == nil || prior.Stage != "BODY_PREFLIGHT" || len(prior.SearchValidations) != 1 || prior.Model != nil ||
		!prior.SearchValidations[0].Observation.Completed || len(prior.SearchValidations[0].Observation.LocalCalls.Evaluations) != 2 {
		t.Fatal("later invalid body erased the original preflight prefix", prior, err)
	}
}
