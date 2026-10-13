package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceBatchPlanner8297(t *testing.T) {
	root := os.Getenv("GOOO_BATCH_PLANNER_SOURCE")
	if root == "" {
		root = filepath.Join("..", "..", "examples", "batch-planner")
	}
	source, err := os.ReadFile(filepath.Join(root, "main.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	built, err := GenerateComposition(context.Background(), "batch-planner.gooo", source, cases, "")
	if err != nil {
		t.Fatal("construct batch planner", err)
	}
	if len(built.Steps) != 1 {
		t.Fatal("one planner activity required", len(built.Steps))
	}
	fill := built.Steps[0].Generation.Report.BodyFill
	if fill == nil || fill.SelectedCandidateID != "inclusive" || fill.TestCasesPassed != 3 || fill.TestCasesTotal != 3 || fill.HoldoutCasesPassed != 1 || fill.HoldoutCasesTotal != 1 || fill.LocalModelPredictions != 0 || fill.ExternalProviderCalls != 0 {
		t.Fatal("source selection and separate holdout", fill)
	}
	run, err := ExecuteComposition(context.Background(), "batch-planner.gooo", source, built, cases, nativeTool())
	if err != nil || run.FinitePassed != 4 || run.FiniteTotal != 4 || run.ModelCalls != 0 {
		t.Fatal("native planner outputs", err, run.FinitePassed, run.FiniteTotal)
	}
	if len(run.ConditionSites) != 3 {
		t.Fatal("planner condition sites", len(run.ConditionSites))
	}
	for _, site := range run.ConditionSites {
		if site.Source == nil || site.Source.Status != "MAPPED" {
			t.Fatal("condition original source", site)
		}
	}
	if output := os.Getenv("GOOO_BATCH_PLANNER_OBSERVATION"); output != "" {
		data, err := json.MarshalIndent(map[string]any{"source": string(source), "cases": json.RawMessage(raw), "construction": built, "runtime": run, "selection_cases": 3, "holdout_cases": 1, "native_cases": 4, "historical_source_executions": 0, "model_calls": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "batch8297.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("8297: source-declared inclusive budget choice selection3/3 holdout1/1; disjoint native4/4 including order-dependent early return, invalid budget and MaxInt64; three original condition sites; model0")
}
