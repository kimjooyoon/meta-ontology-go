package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRuntimeReplaysOracleAndKeepsAddedInputInSelectionSuite(t *testing.T) {
	t.Run("fresh", func(t *testing.T) { runtimeReplaysOracle(t, false) })
	t.Run("reuse", func(t *testing.T) { runtimeReplaysOracle(t, true) })
}

func runtimeReplaysOracle(t *testing.T, reuse bool) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/path-observation.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/path-observation-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := DecodePlan(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := bodycodegen.GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "",
		bodycodegen.TypedPathOptions{Observation: &bodycodegen.PathObservationOptions{
			Inputs: []int64{2, 3, 0}, MaxCandidates: 2, MaxRounds: 2, OracleActivity: "Expected", ReuseProbeOutputs: reuse}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	prior, parent, err := DecodeGeneration(raw)
	if err != nil {
		t.Fatal(err)
	}
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	cases := []pathplan.TestCase{{Input: 2, Expected: 0}, {Input: 3, Expected: -1}, {Input: 0, Expected: 2}}
	result, err := Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, tool)
	if err != nil {
		t.Fatal(err)
	}
	verifyReceipt(t, result)
	d := dimension(t, result, "runtime_selection_disjointness")
	if d.Numerator != 1 || d.Denominator != 3 || !result.Observation.ProjectionReplayed ||
		!result.Observation.RuntimeReplayed || dimension(t, result, "runtime_finite_accuracy").Numerator != 3 {
		t.Fatal("oracle input lost selection identity or failed execution", result.Observation, d)
	}
}
