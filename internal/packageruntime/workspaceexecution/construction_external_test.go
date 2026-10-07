package workspaceexecution

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func mixedConstructionResult(t *testing.T) Result {
	t.Helper()
	manifest, suite := workspaceInputsFixture(t)
	source := &manifest.Packages[0].Sources[0].Content
	*source = strings.Replace(*source, "output = 0", "output = __GOOO_BODY_HOLE_floor__", 1)
	plan := bodycodegen.IRBodyFillPlan{Schema: "gooo/body-codegen-ir-fill-plan/v1", Intent: "Clamp negative inputs to zero.", HoleID: "floor",
		Candidates:       []bodycodegen.IRBodyFillCandidate{{ID: "zero", Expression: "0"}, {ID: "one", Expression: "1"}},
		TestCases:        []bodycodegen.IRBodyFillTestCase{{Input: -2, Expected: 0}, {Input: 0, Expected: 0}, {Input: 7, Expected: 7}},
		HoldoutTestCases: []bodycodegen.IRBodyFillTestCase{{Input: -100, Expected: 0}}}
	result, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite,
		ExecuteOptions{BodyFillPlans: map[string]bodycodegen.IRBodyFillPlan{"example/construction:Normalize": plan}})
	if err != nil || result.Runtime.FinitePassed != 8 {
		t.Fatal("mixed construction did not execute", err, result.Runtime)
	}
	return savedWorkspaceResult(t, result)
}

func TestConstructionObservesExternalAndSourceFillSets(t *testing.T) {
	prior := mixedConstructionResult(t)
	prior.Runtime.FinitePassed, prior.Runtime.FiniteTotal = 999, 999
	rows, err := ObserveConstruction(context.Background(), prior)
	if err != nil || len(rows) != 4 {
		t.Fatal("mixed plan origins were not reconstructed", err, rows)
	}
	for i, row := range rows {
		profile := "external_fill"
		if i >= 2 {
			profile = "source_fill"
		}
		if row.Profile != profile || row.View != "scored_set" || !row.ScoringCompleted || row.Counts.Total != 3 || row.Counts.Best != 3 || row.Counts.Scored != 2 || row.Counts.Budget != 2 {
			t.Fatal("construction scope includes holdout or historical runtime scores", i, row)
		}
	}
	if rows[0].Counts.Matched != 3 || rows[1].Counts.Matched != 2 || rows[2].Counts.Matched != 3 || rows[3].Counts.Matched != 0 {
		t.Fatal("candidate whole-case scores differ", rows)
	}
}

func TestConstructionExternalPlanRequiresMatchingRetainedContract(t *testing.T) {
	prior := mixedConstructionResult(t)
	for _, change := range []string{"missing-plan", "expression", "case", "score", "source-plan", "activity", "order"} {
		t.Run(change, func(t *testing.T) {
			bad := savedWorkspaceResult(t, prior)
			switch change {
			case "missing-plan":
				bad.BodyFills[0].Plan = nil
			case "expression":
				bad.BodyFills[0].Plan.Candidates[0].Expression = "1"
			case "case":
				bad.BodyFills[0].Plan.TestCases[0].Expected = 1
			case "score":
				bad.BodyFills[0].Generation.Report.BodyFill.CandidateScores[1].TestCasesPassed = 3
			case "source-plan":
				bad.BodyFills[1].Plan = bad.BodyFills[0].Plan
			case "activity":
				bad.BodyFills[0].Activity.PackagePath = "example/other"
			case "order":
				bad.BodyFills[0], bad.BodyFills[1] = bad.BodyFills[1], bad.BodyFills[0]
			}
			if _, err := ObserveConstruction(context.Background(), bad); err == nil {
				t.Fatal("changed construction produced policy inputs", change)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ObserveConstruction(ctx, prior); !errors.Is(err, context.Canceled) {
		t.Fatal("construction observation ignored cancellation", err)
	}
}
