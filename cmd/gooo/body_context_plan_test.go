package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestBodyContextExpandedPlanIsExplicitAndSourceBound(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/path-recipe.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	recipe, err := os.ReadFile("../../examples/body-codegen/path-recipe.json")
	if err != nil {
		t.Fatal(err)
	}
	reader := mapSourceReader{"f.gooo": source, "p.json": recipe}
	args := []string{"--plan", "p.json", "--activity", "Compose", "f.gooo"}
	var out, stderr bytes.Buffer
	if runBodyContext(args, reader, &out, &stderr) != exitOK || strings.Contains(out.String(), `"expanded_plan"`) {
		t.Fatal("default export disclosed the plan", out.String(), stderr.String())
	}
	args = append([]string{"--include-plan"}, args...)
	out.Reset()
	if runBodyContext(args, reader, &out, &stderr) != exitOK {
		t.Fatal(out.String(), stderr.String())
	}
	var exported bodyContextOutput
	if err := json.Unmarshal(out.Bytes(), &exported); err != nil {
		t.Fatal(err)
	}
	if exported.ExpandedPlan == nil || !exported.SourceBinding.Equivalent || exported.Context == nil ||
		exported.ModelPredictions != 0 || exported.CandidateTests != 0 || exported.SelectedEmission || exported.RepositoryWrites != 0 {
		t.Fatal("plan was not a source-bound observation")
	}
	planBytes, _ := json.Marshal(exported.ExpandedPlan)
	hash := sha256.Sum256(planBytes)
	if hex.EncodeToString(hash[:]) != exported.Context.OriginalPlanSHA || strings.Contains(string(planBytes), "test_cases") {
		t.Fatal("exported plan identity differs or includes outcomes")
	}
	// Re-validate the exported plan independently of the context output.
	if _, err := pathplan.Prepare(*exported.ExpandedPlan); err != nil {
		t.Fatal(err)
	}
	var recipeFields map[string]any
	if err := json.Unmarshal(recipe, &recipeFields); err != nil {
		t.Fatal(err)
	}
	recipeFields["test_cases"] = []map[string]int64{{"input": 100, "expected": 12345}}
	reader["p.json"], _ = json.Marshal(recipeFields)
	out.Reset()
	if runBodyContext(args, reader, &out, &stderr) != exitOK {
		t.Fatal(out.String(), stderr.String())
	}
	var changed bodyContextOutput
	if err := json.Unmarshal(out.Bytes(), &changed); err != nil {
		t.Fatal(err)
	}
	changedBytes, _ := json.Marshal(changed.ExpandedPlan)
	if !bytes.Equal(planBytes, changedBytes) || changed.TestSuiteSHA256 == exported.TestSuiteSHA256 {
		t.Fatal("case changes affected the source-derived plan")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out.Reset()
	if runBodyContextWithContext(ctx, args, reader, &out, &stderr) != exitFailure || strings.Contains(out.String(), `"expanded_plan"`) {
		t.Fatal("cancelled export disclosed a plan")
	}
}

func TestBodyContextPlanRejectsUnboundDocumentAndDuplicateFlag(t *testing.T) {
	source, _ := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	plan, _ := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	var document pathplan.Document
	if err := json.Unmarshal(plan, &document); err != nil {
		t.Fatal(err)
	}
	for i := range document.Plan.Base.Expressions {
		e := &document.Plan.Base.Expressions[i]
		if e.Kind == "int" {
			e.Int = 999
		}
	}
	changed, _ := json.Marshal(document)
	args := []string{"--include-plan", "--plan", "p", "--activity", "Combined", "f.gooo"}
	var out, stderr bytes.Buffer
	if runBodyContext(args, mapSourceReader{"f.gooo": source, "p": changed}, &out, &stderr) != exitFailure ||
		strings.Contains(out.String(), `"expanded_plan"`) {
		t.Fatal("unbound plan exported", out.String())
	}
	out.Reset()
	if runBodyContext(append(args, "--include-plan"), mapSourceReader{}, &out, &stderr) != exitUsage || out.Len() != 0 {
		t.Fatal("duplicate flag accepted")
	}
}
