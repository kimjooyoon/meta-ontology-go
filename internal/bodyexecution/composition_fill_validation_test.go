package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestCompositionRetainsFillValidationBoundaries(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-calls/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	cases := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":10}}]`)
	prior, err := GenerateCompositionWithOptions(context.Background(), "validation.gooo", source, cases,
		CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(prior)
	var observed struct {
		Checks []json.RawMessage `json:"fill_validations"`
	}
	if err := json.Unmarshal(raw, &observed); err != nil || len(observed.Checks) != 2 {
		t.Fatal("original fill preflight boundaries were discarded", len(observed.Checks), err)
	}
	for i, stage := range []string{"BODY_PREFLIGHT", "ACTIVITY_GENERATION"} {
		check := prior.FillValidations[i]
		if check.Stage != stage || !check.Observation.Completed || check.Observation.SourceSHA256 != digest(source) ||
			check.Observation.LocalCalls == nil || len(check.Observation.LocalCalls.Evaluations) != 6 {
			t.Fatal("validation checkpoint or original passes differ", check)
		}
	}
	raw, _ = json.MarshalIndent(prior, "", "  ")
	saved, err := DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := VerifyComposition(ctx, "validation.gooo", source, saved); err != nil {
		t.Fatal("saved fill validation did not replay", err)
	}
	runtime, err := ExecuteComposition(ctx, "validation.gooo", source, saved, cases, nativeTool())
	if err != nil || runtime.FinitePassed != 1 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("fill validation did not reach native execution", runtime, err)
	}
	for name, mutate := range map[string]func(*Composition){
		"stage":  func(c *Composition) { c.FillValidations[0].Stage = "ACTIVITY_GENERATION" },
		"source": func(c *Composition) { c.FillValidations[1].Observation.SourceSHA256 = "changed" },
		"argument": func(c *Composition) {
			c.FillValidations[0].Observation.LocalCalls.Evaluations[0].Calls[0].Inputs = json.RawMessage("[99]")
		},
		"missing": func(c *Composition) { c.FillValidations = c.FillValidations[:1] },
		"empty":   func(c *Composition) { c.FillValidations = []CompositionFillValidation{} },
		"extra":   func(c *Composition) { c.FillValidations = append(c.FillValidations, c.FillValidations[0]) },
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
	saved.FillValidations = nil
	if err := VerifyComposition(ctx, "validation.gooo", source, saved); err != nil || saved.FillValidations != nil {
		t.Fatal("historical absence was rewritten", err)
	}
}

func TestCompositionFillValidationUsesPreparedHelperCheckpoint(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-history/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	cases := jointCases(t, `[{"inputs":{"Main":10},"expected":{"Main":5}}]`)
	prior, err := GenerateCompositionWithOptions(context.Background(), "prepared.gooo", source, cases,
		CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.FillValidations) != 1 || prior.FillValidations[0].Stage != "ACTIVITY_GENERATION" ||
		prior.FillValidations[0].Observation.Activity != "Build" ||
		prior.FillValidations[0].Observation.SourceSHA256 == digest(source) ||
		prior.FillValidations[0].Observation.SourceSHA256 != prior.Preparations[1].InputSourceSHA256 {
		t.Fatal("deferred fill validation lost its prepared source", prior.FillValidations)
	}
	if err := VerifyComposition(context.Background(), "prepared.gooo", source, prior); err != nil {
		t.Fatal("prepared source fill validation did not replay", err)
	}
}

func TestCompositionFillValidationKeepsEarlierPreflightBeforeFailure(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-calls/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, []byte(`
activity Broken(Integer) -> Integer computes "return missing"
bind Build.result -> Broken.input
`)...)
	cases := jointCases(t, `[{"inputs":{"Build":4},"expected":{"Broken":10}}]`)
	prior, err := GenerateCompositionWithOptions(context.Background(), "partial.gooo", source, cases,
		CompositionOptions{EntryActivity: "Broken", FillModelPath: "missing-model.json"})
	if err == nil || prior.Stage != "BODY_PREFLIGHT" || len(prior.FillValidations) != 1 || prior.FillModel != nil ||
		!prior.FillValidations[0].Observation.Completed || !strings.Contains(err.Error(), "Broken") {
		t.Fatal("later invalid body erased earlier preflight or loaded a model", prior, err)
	}
}

func TestCompositionFillValidationInputsKeepPositiveExposureAndUnknown(t *testing.T) {
	graph, seen := localInputFixture()
	prior := Composition{FillValidations: []CompositionFillValidation{{Stage: "BODY_PREFLIGHT",
		Observation: bodycodegen.SourceFillValidation{ActivityID: "root", LocalCalls: &bodycodegen.LocalCallHistory{
			Schema: "gooo/local-call-inputs/v1", Status: "RECORDED",
			Evaluations: []*bodycodegen.LocalCallObservation{localInputObservation("seed", "[9007199254740993]")},
		}},
	}}}
	if err := graph.includeCompositionLocalInputs(&seen, prior); err != nil {
		t.Fatal(err)
	}
	if _, ok := seen.called["seed"]["[9007199254740993]"]; !ok || !seen.unobservedConstructionCalls {
		t.Fatal("preflight exposure disappeared or became full coverage")
	}
}
