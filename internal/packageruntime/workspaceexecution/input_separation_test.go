package workspaceexecution

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func workspaceInputsFixture(t *testing.T) (packageruntime.Manifest, bodyexecution.CompositionCases) {
	t.Helper()
	root := "../../../examples/workspace-input-observations/"
	source, err := os.ReadFile(root + "main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	cases, err := os.ReadFile(root + "cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/construction", Activity: "Lift"},
		Packages: []packageruntime.PackageSpec{{Path: "example/construction", Name: "construction",
			Sources: []packageruntime.Source{{Filename: "main.gooo", Content: string(source)}}}}}
	return manifest, suite
}

func TestWorkspaceInputSeparationIncludesEarlierSourceFill(t *testing.T) {
	manifest, suite := workspaceInputsFixture(t)
	for _, mode := range []string{"pass", "duplicate-failure", "inputs-only"} {
		t.Run(mode, func(t *testing.T) {
			current := suite
			current.Cases = append([]bodyexecution.CompositionCase(nil), suite.Cases...)
			switch mode {
			case "duplicate-failure":
				current.Cases[3].Expected = maps.Clone(current.Cases[3].Expected)
				current.Cases[3].Expected["example/construction:Lift"] = json.RawMessage("999")
			case "inputs-only":
				current.Schema = bodyexecution.CompositionInputsSchema
				for i := range current.Cases {
					current.Cases[i].Expected = nil
				}
			}
			result, err := ExecuteWorkspace(context.Background(), manifest, current, "", "")
			if err != nil {
				t.Fatal(err)
			}
			checkWorkspaceInputCounts(t, result.Runtime.InputSeparation, mode)
		})
	}
}

func checkWorkspaceInputCounts(t *testing.T, got bodyexecution.CompositionInputSeparation, mode string) {
	t.Helper()
	if got.UniqueInputs != 3 || got.DuplicateRows != 1 || got.OverlappingInputs != 2 || got.DisjointInputs != 1 || got.UnknownInputs != 0 || len(got.EarlierStages) != 1 {
		t.Fatal("earlier training/holdout input or duplicate accounting differs", got)
	}
	wantStatus, wantPassed := "PASS", 1
	if mode == "duplicate-failure" {
		wantStatus, wantPassed = "PROGRESS", 0
	} else if mode == "inputs-only" {
		wantStatus, wantPassed = "UNKNOWN", 0
	}
	if got.Status != wantStatus || got.DisjointCasesPassed != wantPassed {
		t.Fatal("finite outcome or missing expectations differ", mode, got)
	}
	stage := got.EarlierStages[0]
	if stage.ActivityID == "" || stage.SourceSHA256 == "" || stage.PlanSHA256 == "" || stage.InputSetSHA256 == "" || stage.UniqueInputs < 4 {
		t.Fatal("earlier-stage input identity is missing", stage)
	}
}

func TestWorkspaceInputSeparationIncludesExternalFill(t *testing.T) {
	manifest, suite := workspaceInputsFixture(t)
	source := &manifest.Packages[0].Sources[0].Content
	start, end := strings.Index(*source, "activity Lift"), strings.Index(*source, "\n\nbind")
	*source = (*source)[:start] + "activity Lift(Integer) -> Integer computes \"return __GOOO_BODY_HOLE_value__\"" + (*source)[end:]
	plan := bodycodegen.IRBodyFillPlan{Schema: "gooo/body-codegen-ir-fill-plan/v1", Intent: "Add one.", HoleID: "value",
		Candidates:       []bodycodegen.IRBodyFillCandidate{{ID: "one", Expression: "input + 1"}, {ID: "double", Expression: "input * 2"}},
		TestCases:        []bodycodegen.IRBodyFillTestCase{{Input: -2, Expected: -1}, {Input: 0, Expected: 1}, {Input: 7, Expected: 8}},
		HoldoutTestCases: []bodycodegen.IRBodyFillTestCase{{Input: 100, Expected: 101}}}
	result, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite,
		ExecuteOptions{BodyFillPlans: map[string]bodycodegen.IRBodyFillPlan{"example/construction:Lift": plan}})
	if err != nil {
		t.Fatal(err)
	}
	checkWorkspaceInputCounts(t, result.Runtime.InputSeparation, "pass")
}
