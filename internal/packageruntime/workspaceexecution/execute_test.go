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

func TestExecuteWorkspaceCarriesTypedDomainRecordsAcrossBindings(t *testing.T) {
	manifest := packageruntime.Manifest{
		Schema: packageruntime.ManifestSchema,
		Entry:  packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "example/app", Name: "app", Imports: []string{"example/domain"}, Sources: []packageruntime.Source{{
				Filename: "app.gooo", Content: `package app
namespace app
import domain "example/domain"
activity Main(Candidate) -> Review computes ` + "`" + `return Review{candidate_id: input.candidate_id, decision: "accepted"}` + "`" + `
bind domain.Submit.result -> Main.input
`,
			}}},
			{Path: "example/domain", Name: "domain", Sources: []packageruntime.Source{{
				Filename: "domain.gooo", Content: `package domain
namespace domain
entity Candidate id "example://domain/candidate" fields {
    field candidate_id id "example://domain/candidate/id" type integer required one
    field title id "example://domain/candidate/title" type string required one
}
entity Review id "example://domain/review" fields {
    field candidate_id id "example://domain/review/candidate-id" type integer required one
    field decision id "example://domain/review/decision" type string required one
}
activity Submit(Candidate) -> Candidate computes "return input"
`,
			}}},
		},
	}
	input, _ := json.Marshal(map[string]any{"candidate_id": 41, "title": "queue"})
	review, _ := json.Marshal(map[string]any{"candidate_id": 41, "decision": "accepted"})
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs: map[string]json.RawMessage{"example/domain:Submit": input},
		Expected: map[string]json.RawMessage{
			"example/domain:Submit": input,
			"example/app:Main":      review,
		},
	}}}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatalf("execute typed record activity graph: %v (stage %s, failure %s)", err, result.Runtime.Stage, result.Runtime.Failure)
	}
	if result.Runtime.Stage != "COMPLETE" || !result.Runtime.ProjectionReplayed || !result.Runtime.RuntimeReplayed ||
		result.Runtime.FinitePassed != 2 || result.Runtime.FiniteTotal != 2 || len(result.Runtime.Traces) != 1 ||
		len(result.Runtime.Traces[0].Deliveries) != 2 {
		t.Fatalf("typed record flow evidence is incomplete: %+v", result.Runtime)
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
			step.Generation.Report.BodyFill.Timing.DecisionStage != "after_typed_ir_plan_training_scores_and_behavior_probes_before_final_emission" {
			t.Fatalf("model body-fill provenance or accuracy missing: %#v", step)
		}
	}
}

func TestExecuteWorkspaceUsesSourceDeclaredFillPlanWithLayaAndDeterministicFallback(t *testing.T) {
	manifest := packageruntime.Manifest{
		Schema: packageruntime.ManifestSchema,
		Entry:  packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "example/app", Name: "app", Imports: []string{"example/core"}, Sources: []packageruntime.Source{{
				Filename: "app.gooo", Content: `package app
namespace app
import core "example/core"
activity Main(Integer) -> Integer computes "return input"
bind core.Normalize.result -> Main.input
`,
			}}},
			{Path: "example/core", Name: "core", Sources: []packageruntime.Source{{
				Filename: "core.gooo", Content: `package core
namespace core
entity Integer id "example://core/integer"
activity Normalize(Integer) -> Integer computes ` + "`" + `let base = __GOOO_BODY_HOLE_seed__
let increment = __GOOO_BODY_HOLE_step__
return base + increment` + "`" + ` assembling {
    source_fill intent "Represent input plus one as a base value and a step." {
        hole "seed"
        hole "step"
        candidate "add_one" {
            fill "seed" "input + 0"
            fill "step" "1"
        }
        candidate "subtract_zero_then_add" {
            fill "seed" "input - 0"
            fill "step" "1"
        }
    }
    case "-4" -> "-3"
    case "0" -> "1"
    case "7" -> "8"
}


`,
			}}},
		},
	}
	seven, _ := json.Marshal(7)
	eight, _ := json.Marshal(8)
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs:   map[string]json.RawMessage{"example/core:Normalize": seven},
		Expected: map[string]json.RawMessage{"example/core:Normalize": eight, "example/app:Main": eight},
	}}}
	var calls atomic.Int32
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
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" ||
			json.NewDecoder(request.Body).Decode(&input) != nil {
			http.Error(writer, "invalid choice request", http.StatusBadRequest)
			return
		}
		criteria := input.Questions["body_ir_fill"].Criteria
		if _, ok := criteria["subtract_zero_then_add"]; !ok {
			http.Error(writer, "source-declared assignment missing", http.StatusBadRequest)
			return
		}
		calls.Add(1)
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "fixture-model", "routing": map[string]any{"model": "fixture-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{"choice": "subtract_zero_then_add"}},
		})
	}))
	defer server.Close()
	modelResult, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite, ExecuteOptions{LayaEndpoint: server.URL + "/v1/systemone"})
	if err != nil {
		t.Fatalf("execute Gooo source-declared body fill through Laya: %v", err)
	}
	if calls.Load() != 1 || len(modelResult.BodyFills) != 1 || modelResult.BodyFills[0].Generation.Report.BodyFill.SelectedCandidateID != "subtract_zero_then_add" ||
		modelResult.BodyFills[0].Generation.Report.BodyFill.Decision.Provider != "laya" || modelResult.Runtime.FinitePassed != 2 ||
		modelResult.Runtime.FiniteTotal != 2 || !modelResult.Runtime.RuntimeReplayed || strings.Contains(modelResult.BodyFills[0].Generation.GoooSource, "source_fill") {
		t.Fatalf("Gooo source plan was not consumed, selected, and executed: calls=%d fills=%+v runtime=%+v", calls.Load(), modelResult.BodyFills, modelResult.Runtime)
	}
	_, err = ExecuteWorkspaceWithOptions(context.Background(), manifest, suite, ExecuteOptions{
		BodyFillPlans: map[string]bodycodegen.IRBodyFillPlan{"example/core:Normalize": {}},
	})
	if err == nil || !strings.Contains(err.Error(), "both a Gooo source fill plan and an external body-fill plan") {
		t.Fatalf("source and external plan ambiguity did not fail closed: %v", err)
	}
	fallbackResult, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite, ExecuteOptions{})
	if err != nil {
		t.Fatalf("execute Gooo source-declared body fill without a model: %v", err)
	}
	if len(fallbackResult.BodyFills) != 1 || fallbackResult.BodyFills[0].Generation.Report.BodyFill.SelectedCandidateID != "add_one" ||
		fallbackResult.BodyFills[0].Generation.Report.BodyFill.Decision.Provider != "deterministic" ||
		fallbackResult.Runtime.FinitePassed != 2 || fallbackResult.Runtime.FiniteTotal != 2 || !fallbackResult.Runtime.RuntimeReplayed {
		t.Fatalf("Gooo source plan did not use deterministic fallback: fills=%+v runtime=%+v", fallbackResult.BodyFills, fallbackResult.Runtime)
	}
}

func TestExecuteWorkspaceUsesLayaForTypedRecordBodyFill(t *testing.T) {
	manifest := packageruntime.Manifest{
		Schema: packageruntime.ManifestSchema,
		Entry:  packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "example/app", Name: "app", Imports: []string{"example/domain"}, Sources: []packageruntime.Source{{
				Filename: "app.gooo", Content: `package app
namespace app
import domain "example/domain"
activity Main(Candidate) -> Review computes ` + "`" + `if __GOOO_BODY_HOLE_condition__ {
    return Review{candidate_id: input.candidate_id, decision: __GOOO_BODY_HOLE_accepted__}
} else {
    return Review{candidate_id: input.candidate_id, decision: __GOOO_BODY_HOLE_rejected__}
}` + "`" + ` assembling {
    source_fill intent "Accept ready candidates and reject all others." {
        hole "condition"
        hole "accepted"
        hole "rejected"
        candidate "ready_is_accepted" {
            fill "condition" "input.state == \"ready\""
            fill "accepted" "\"accepted\""
            fill "rejected" "\"rejected\""
        }
        candidate "ready_is_rejected" {
            fill "condition" "input.state != \"ready\""
            fill "accepted" "\"accepted\""
            fill "rejected" "\"rejected\""
        }
    }
    value_case "[{\"candidate_id\":41,\"state\":\"ready\"}]" -> "{\"candidate_id\":41,\"decision\":\"accepted\"}"
    value_case "[{\"candidate_id\":42,\"state\":\"queued\"}]" -> "{\"candidate_id\":42,\"decision\":\"rejected\"}"
}
bind domain.Submit.result -> Main.input
`,
			}}},
			{Path: "example/domain", Name: "domain", Sources: []packageruntime.Source{{
				Filename: "domain.gooo", Content: `package domain
namespace domain
entity Candidate id "example://domain/candidate" fields {
    field candidate_id id "example://domain/candidate/id" type integer required one
    field state id "example://domain/candidate/state" type string required one
}
entity Review id "example://domain/review" fields {
    field candidate_id id "example://domain/review/candidate-id" type integer required one
    field decision id "example://domain/review/decision" type string required one
}
activity Submit(Candidate) -> Candidate computes "return input"
`,
			}}},
		},
	}
	ready, _ := json.Marshal(map[string]any{"candidate_id": 41, "state": "ready"})
	queued, _ := json.Marshal(map[string]any{"candidate_id": 42, "state": "queued"})
	accept, _ := json.Marshal(map[string]any{"candidate_id": 41, "decision": "accepted"})
	reject, _ := json.Marshal(map[string]any{"candidate_id": 42, "decision": "rejected"})
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{
		{Inputs: map[string]json.RawMessage{"example/domain:Submit": ready}, Expected: map[string]json.RawMessage{"example/domain:Submit": ready, "example/app:Main": accept}},
		{Inputs: map[string]json.RawMessage{"example/domain:Submit": queued}, Expected: map[string]json.RawMessage{"example/domain:Submit": queued, "example/app:Main": reject}},
	}}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/health" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"revisions": map[string]string{"record-model": "0123456789abcdef0123456789abcdef"}})
			return
		}
		var payload struct {
			State     map[string]string `json:"state"`
			Questions map[string]struct {
				Criteria map[string]string `json:"criteria"`
			} `json:"questions"`
		}
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" ||
			json.NewDecoder(request.Body).Decode(&payload) != nil {
			http.Error(writer, "invalid record-fill choice request", http.StatusBadRequest)
			return
		}
		var state struct {
			InputType  string `json:"input_type"`
			OutputType string `json:"output_type"`
			TestCount  int    `json:"test_case_count"`
		}
		if json.Unmarshal([]byte(payload.State["request"]), &state) != nil || state.InputType != "Candidate" ||
			state.OutputType != "Review" || state.TestCount != 2 || len(payload.Questions["body_ir_fill"].Criteria) != 2 {
			http.Error(writer, "typed record IR missing from model request", http.StatusBadRequest)
			return
		}
		calls.Add(1)
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "record-model", "routing": map[string]any{"model": "record-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{
				"choice": "ready_is_accepted", "probabilities": map[string]float64{"ready_is_accepted": 1},
			}},
		})
	}))
	defer server.Close()

	result, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite, ExecuteOptions{LayaEndpoint: server.URL + "/v1/systemone"})
	if err != nil {
		t.Fatalf("execute model-composed record activity graph: %v", err)
	}
	if calls.Load() != 1 || len(result.BodyFills) != 1 || result.BodyFills[0].Generation.Report.BodyFill == nil ||
		result.BodyFills[0].Generation.Report.BodyFill.SelectedCandidateID != "ready_is_accepted" ||
		result.BodyFills[0].Generation.Report.BodyFill.Decision.Provider != "laya" ||
		result.BodyFills[0].Generation.Report.BodyFill.FunctionalAccuracyPct != 100 ||
		result.Runtime.FinitePassed != 4 || result.Runtime.FiniteTotal != 4 || !result.Runtime.RuntimeReplayed {
		t.Fatalf("model-composed record execution did not retain generation and runtime evidence: calls=%d selected=%+v runtime=%+v",
			calls.Load(), result.BodyFills[0].Generation.Report.BodyFill, result.Runtime)
	}
}
