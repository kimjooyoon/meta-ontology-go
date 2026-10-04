package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func compositionFixture(t *testing.T) ([]byte, CompositionCases) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/native-composition.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/native-composition-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, suite
}

func TestCompositionTypedSourceGraphAndReplay(t *testing.T) {
	source, suite := compositionFixture(t)
	prior, err := GenerateComposition(context.Background(), "graph.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Steps) != 7 || len(prior.Plan.Edges) != 5 || prior.Model == nil || prior.Model.Loaded {
		t.Fatalf("composition coverage: %+v", prior.Plan)
	}
	if strings.Count(prior.GoooSource, "baseline ") != 2 || strings.Count(prior.GoooSource, "picked ") != 6 {
		t.Fatal("both independently assembled checkpoints were not retained")
	}
	if err := VerifyComposition(context.Background(), "graph.gooo", source, prior); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyComposition(context.Background(), "graph.gooo", source, decoded); err != nil {
		t.Fatal(err)
	}
	next, err := GenerateComposition(context.Background(), "graph.gooo", []byte(prior.GoooSource), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if next.Source != prior.Source || next.GoooSource != prior.GoooSource {
		t.Fatal("composed checkpoints did not reach a source/code fixed point")
	}
}

func TestCompositionNativeMultiTypeFanoutAndPartialScores(t *testing.T) {
	source, suite := compositionFixture(t)
	prior, err := GenerateComposition(context.Background(), "graph.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "graph.gooo", source, prior, suite, nativeTool())
	if err != nil {
		t.Fatalf("native graph: %v; %+v", err, run)
	}
	if !run.ProjectionReplayed || !run.RuntimeReplayed || run.FinitePassed != 49 || run.FiniteTotal != 49 ||
		len(run.Runs) != 2 || len(run.Traces) != 7 || run.ModelCalls != 0 {
		t.Fatalf("native observation coverage: %+v", run)
	}
	for _, trace := range run.Traces {
		for i, step := range trace.Deliveries {
			node := prior.Plan.Activities[i]
			if node.InputFrom >= 0 && (!bytes.Equal(step.Input, trace.Deliveries[node.InputFrom].Actual) ||
				step.ProducerID != prior.Plan.Activities[node.InputFrom].ID) {
				t.Fatal("bound input differs from native producer output")
			}
		}
	}
	suite.Cases[0].Expected["Clamp"] = json.RawMessage("999")
	partial, err := ExecuteComposition(context.Background(), "graph.gooo", source, prior, suite, nativeTool())
	if err != nil || !partial.RuntimeReplayed || partial.FinitePassed != 48 || partial.FiniteTotal != 49 {
		t.Fatalf("partial finite score discarded: %v %+v", err, partial)
	}
}

func TestCompositionRejectsGraphAndCaseDefectsBeforeModel(t *testing.T) {
	source, suite := compositionFixture(t)
	for _, change := range []string{"cycle", "multiple", "type", "missing-edge", "bound-input", "missing-root", "unknown-result", "null", "fraction", "unsupported-body"} {
		t.Run(change, func(t *testing.T) {
			raw, _ := json.Marshal(suite)
			cases, _ := DecodeCompositionCases(raw)
			candidate := string(source)
			switch change {
			case "cycle":
				candidate += "\nbind Twice.result -> Assemble.input\n"
			case "multiple":
				candidate += "\nbind Twice.result -> Clamp.input\n"
			case "type":
				candidate = strings.Replace(candidate, "bind Clamp.result -> IsPositive.input", "bind Invert.result -> IsPositive.input", 1)
			case "missing-edge":
				candidate = strings.Replace(candidate, "bind Assemble.result -> Clamp.input", "", 1)
			case "bound-input":
				cases.Cases[0].Inputs["Clamp"] = json.RawMessage("0")
			case "missing-root":
				delete(cases.Cases[0].Inputs, "Decorate")
			case "unknown-result":
				cases.Cases[0].Expected["Unknown"] = json.RawMessage("0")
			case "null":
				cases.Cases[0].Inputs["Assemble"] = json.RawMessage("null")
			case "fraction":
				cases.Cases[0].Inputs["Assemble"] = json.RawMessage("1.0")
			case "unsupported-body":
				candidate = strings.Replace(candidate, "return input * 2", "return missing(input)", 1)
			}
			result, err := GenerateComposition(context.Background(), "invalid.gooo", []byte(candidate), cases, "missing-model.json")
			if err == nil || strings.Contains(err.Error(), "load retained") || result.Model != nil || len(result.Steps) != 0 {
				t.Fatalf("defect reached model/generation: %v %+v", err, result)
			}
		})
	}
}

func TestCompositionReplayRejectsSourceStepAndProjectionChanges(t *testing.T) {
	source, suite := compositionFixture(t)
	prior, err := GenerateComposition(context.Background(), "graph.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(prior)
	for _, change := range []string{"source", "plan", "order", "input", "generated", "driver", "selected", "body", "cases", "plain"} {
		t.Run(change, func(t *testing.T) {
			candidate, _ := DecodeComposition(encoded)
			original := append([]byte(nil), source...)
			switch change {
			case "source":
				original = append(original, '\n')
			case "plan":
				candidate.Plan.Activities[0].ID += "changed"
			case "order":
				candidate.Steps[0], candidate.Steps[1] = candidate.Steps[1], candidate.Steps[0]
			case "input":
				candidate.Steps[1].InputSourceSHA256 = "other"
			case "generated":
				candidate.Source += "\n"
			case "driver":
				candidate.Driver += "\n"
			case "selected":
				candidate.GoooSource += "\n"
			case "body":
				candidate.Steps[0].Generation.Source += "\n"
			case "cases":
				candidate.Steps[0].Generation.Report.BodyPaths.NativeCases[0].Actual++
			case "plain":
				for i := range candidate.Steps {
					if candidate.Steps[i].Generation.Report.BodyPaths == nil {
						candidate.Steps[i].Generation.Report.ActivityID = "other"
						break
					}
				}
			}
			run, err := ExecuteComposition(context.Background(), "graph.gooo", original, candidate, suite, "missing-tool")
			if err == nil || run.ProjectionReplayed || run.Build.Started || len(run.Runs) != 0 {
				t.Fatalf("invalid composition reached native execution: %v %+v", err, run)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := VerifyComposition(ctx, "graph.gooo", source, prior); err == nil {
		t.Fatal("cancelled replay succeeded")
	}
}

func TestCompositionCaseJSONExactScalarsAndKeys(t *testing.T) {
	for _, raw := range []string{
		`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"A":0,"A":1},"expected":{"B":1}}]}`,
		`{"schema":"gooo/body-composition-cases/v1","Cases":[]}`,
		`{"schema":"gooo/body-composition-cases/v1","cases":[]} {}`,
	} {
		if _, err := DecodeCompositionCases([]byte(raw)); err == nil {
			t.Fatal("invalid JSON accepted", raw)
		}
	}
	for _, pair := range [][2]string{{"Integer", "9223372036854775808"}, {"Boolean", "0"}, {"Text", "true"}} {
		if _, err := canonicalScalar([]byte(pair[1]), pair[0]); err == nil {
			t.Fatal("wrong scalar accepted", pair)
		}
	}
	graph, err := prepareCompositionGraph(context.Background(), "graph.gooo", func() []byte { source, _ := compositionFixture(t); return source }())
	if err != nil {
		t.Fatal(err)
	}
	source, _ := compositionFixture(t)
	reordered := strings.ReplaceAll(string(source), "bind Assemble.result -> Clamp.input\n", "") + "\nbind Assemble.result -> Clamp.input\n"
	other, err := prepareCompositionGraph(context.Background(), "other.gooo", []byte(reordered))
	if err != nil || !reflect.DeepEqual(graph.plan, other.plan) {
		t.Fatalf("edge source order changed deterministic graph: %v", err)
	}
}

func TestCompositionFullArrayBoundAndUnscoredIntermediateValues(t *testing.T) {
	var source strings.Builder
	source.WriteString("package chain\nnamespace chain\nentity Integer id \"chain://integer\"\n")
	for i := range 16 {
		fmt.Fprintf(&source, "activity A%02d(Integer) -> Integer computes \"return input + 1\"\n", i)
	}
	for i := 1; i < 16; i++ {
		fmt.Fprintf(&source, "bind A%02d.result -> A%02d.input\n", i-1, i)
	}
	suite, err := DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"A00":0},"expected":{"A15":16}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "chain.gooo", []byte(source.String()), suite, "")
	if err != nil || len(prior.Steps) != 16 || prior.Model != nil {
		t.Fatalf("16-stage generation: %v", err)
	}
	run, err := ExecuteComposition(context.Background(), "chain.gooo", []byte(source.String()), prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 1 || run.FiniteTotal != 1 || len(run.Traces[0].Deliveries) != 16 {
		t.Fatalf("full-array graph: %v %+v", err, run)
	}
	for i, step := range run.Traces[0].Deliveries {
		if i < 15 && (step.Passed != nil || len(step.Expected) != 0) {
			t.Fatal("unscored intermediate output received accuracy credit")
		}
	}
	oversized := source.String() + "\nactivity Extra(Integer) -> Integer computes \"return input\"\n"
	if _, err := GenerateComposition(context.Background(), "large.gooo", []byte(oversized), suite, ""); err == nil {
		t.Fatal("17-stage graph exceeded fixed bound")
	}
	if failed, err := GenerateComposition(context.Background(), "chain.gooo", []byte(source.String()), suite, "missing-model"); err == nil || len(failed.Steps) != 0 {
		t.Fatal("unused model request was silently ignored")
	}
}
