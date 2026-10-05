package workspaceexecution

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func importedActivityWorkspace() packageruntime.Manifest {
	return packageruntime.Manifest{
		Schema: packageruntime.ManifestSchema,
		Entry:  packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "example/app", Name: "app", Imports: []string{"example/core"}, Sources: []packageruntime.Source{{
				Filename: "app.gooo", Content: `package app
namespace app
import core "example/core"
activity Main(Text) -> Text computes "return input"
bind core.Normalize.result -> Main.input
`,
			}}},
			{Path: "example/core", Name: "core", Sources: []packageruntime.Source{{
				Filename: "core.gooo", Content: `package core
namespace core
entity Text id "example://core/text"
activity Normalize(Text) -> Text computes "return input"
activity Unused(Text) -> Text computes "return input"
`,
			}}},
		},
	}
}

func TestPrepareLowersImportedBindingToTypedWorkspaceGraph(t *testing.T) {
	program, err := Prepare(importedActivityWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	if program.Schema != "gooo/workspace-body-program/v1" || program.Entry.PackagePath != "example/app" || program.Entry.Activity != "Main" {
		t.Fatalf("program identity = %#v", program)
	}
	if len(program.Activities) != 2 || !strings.Contains(program.Source, "bind GoooPackage0ActivityNormalize.result -> GoooPackage2ActivityMain.input") ||
		strings.Contains(program.Source, "ActivityUnused") {
		t.Fatalf("imported binding was not lowered: activities=%#v\n%s", program.Activities, program.Source)
	}
	if !strings.Contains(program.Source, `entity Text id "example://core/text"`) || strings.Contains(program.Source, "import core") {
		t.Fatalf("flattened source lost the stable entity or retained package-only import syntax:\n%s", program.Source)
	}
}

func TestExecuteWorkspaceBuildsAndRunsImportedActivityBodies(t *testing.T) {
	value, _ := json.Marshal("hello")
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs:   map[string]json.RawMessage{"example/core:Normalize": value},
		Expected: map[string]json.RawMessage{"example/core:Normalize": value, "example/app:Main": value},
	}}}
	result, err := ExecuteWorkspace(context.Background(), importedActivityWorkspace(), suite, "", "")
	if err != nil {
		t.Fatalf("execute imported activity graph: %v (stage %s, failure %s)", err, result.Runtime.Stage, result.Runtime.Failure)
	}
	if result.Runtime.Stage != "COMPLETE" || !result.Runtime.ProjectionReplayed || !result.Runtime.RuntimeReplayed ||
		result.Runtime.FinitePassed != 2 || result.Runtime.FiniteTotal != 2 || len(result.Runtime.Traces) != 1 ||
		len(result.Runtime.Traces[0].Deliveries) != 2 {
		t.Fatalf("workspace execution evidence is incomplete: %+v", result.Runtime)
	}
}

func TestPrepareRejectsCrossPackageEntityNameCollision(t *testing.T) {
	manifest := importedActivityWorkspace()
	manifest.Packages[0].Sources = []packageruntime.Source{{Filename: "app.gooo", Content: `package app
namespace app
import core "example/core"
entity Text id "example://app/text"
activity Main(Text) -> Text computes "return input"
`}}
	if _, err := Prepare(manifest); err == nil || !strings.Contains(err.Error(), "different stable IDs") {
		t.Fatalf("expected explicit entity collision rejection; got %v", err)
	}
}

func TestExecuteWorkspaceSelectsGoooBodyFillBeforeNativePackageBuild(t *testing.T) {
	manifest := packageruntime.Manifest{
		Schema: packageruntime.ManifestSchema,
		Entry:  packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "example/app", Name: "app", Imports: []string{"example/core"}, Sources: []packageruntime.Source{{
				Filename: "app.gooo", Content: `package app
namespace app
import core "example/core"
activity Main(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"
bind core.Normalize.result -> Main.input
`,
			}}},
			{Path: "example/core", Name: "core", Sources: []packageruntime.Source{{
				Filename: "core.gooo", Content: `package core
namespace core
entity Integer id "example://core/integer"
activity Normalize(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"
`,
			}}},
		},
	}
	plan := func(intent string, preferredID, preferredExpr, alternateID, alternateExpr string, expected []int64) bodycodegen.IRBodyFillPlan {
		return bodycodegen.IRBodyFillPlan{Schema: "gooo/body-codegen-ir-fill-plan/v1", Intent: intent, HoleID: "value",
			Candidates: []bodycodegen.IRBodyFillCandidate{{ID: preferredID, Expression: preferredExpr}, {ID: alternateID, Expression: alternateExpr}},
			TestCases:  []bodycodegen.IRBodyFillTestCase{{Input: -2, Expected: expected[0]}, {Input: 0, Expected: expected[1]}, {Input: 7, Expected: expected[2]}}}
	}
	plans := map[string]bodycodegen.IRBodyFillPlan{
		"example/core:Normalize": plan("add one", "increment", "input + 1", "identity", "input", []int64{-1, 1, 8}),
		"example/app:Main":       plan("pass through the imported result", "identity", "input", "zero", "0", []int64{-2, 0, 7}),
	}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/health" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"revisions": map[string]string{"fixture-model": "0123456789abcdef0123456789abcdef"}})
			return
		}
		var input struct {
			Questions map[string]struct {
				Criteria map[string]string `json:"criteria"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil || len(input.Questions["body_ir_fill"].Criteria) == 0 {
			http.Error(writer, "invalid choice request", http.StatusBadRequest)
			return
		}
		requests.Add(1)
		choice := "identity"
		if _, ok := input.Questions["body_ir_fill"].Criteria["increment"]; ok {
			choice = "increment"
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "fixture-model", "routing": map[string]any{"model": "fixture-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{"choice": choice}},
		})
	}))
	defer server.Close()
	seven, _ := json.Marshal(7)
	eight, _ := json.Marshal(8)
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs:   map[string]json.RawMessage{"example/core:Normalize": seven},
		Expected: map[string]json.RawMessage{"example/core:Normalize": eight, "example/app:Main": eight},
	}}}
	result, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite, ExecuteOptions{
		BodyFillPlans: plans, LayaEndpoint: server.URL + "/v1/systemone",
	})
	if err != nil {
		t.Fatalf("execute model-selected package bodies: %v", err)
	}
	if requests.Load() != 2 || len(result.BodyFills) != 2 || result.Runtime.FinitePassed != 2 || result.Runtime.FiniteTotal != 2 || !result.Runtime.RuntimeReplayed {
		t.Fatalf("model-selected package execution incomplete: requests=%d fills=%d runtime=%+v", requests.Load(), len(result.BodyFills), result.Runtime)
	}
	if result.BodyFills[1].InputSourceSHA256 != sourceSHA256([]byte(result.BodyFills[0].Generation.GoooSource)) {
		t.Fatal("second Laya body decision did not consume the first selected Gooo source")
	}
	for _, step := range result.BodyFills {
		if step.Generation.Report.BodyFill.Decision.Provider != "laya" || step.Generation.GoooSource == "" ||
			step.Generation.Report.BodyFill.FunctionalAccuracyPct != 100 ||
			step.Generation.Report.BodyFill.Timing.ExecutionModel != "synchronous_sequential_no_background_codegen_goroutines" ||
			step.Generation.Report.BodyFill.Timing.DecisionStage != "after_typed_ir_plan_and_candidate_test_scores_before_final_emission" {
			t.Fatalf("model body-fill provenance or accuracy missing: %#v", step)
		}
	}
}
