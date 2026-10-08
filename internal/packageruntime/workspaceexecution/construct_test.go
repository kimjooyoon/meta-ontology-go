package workspaceexecution

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func callerConstructionFixture(t *testing.T) (packageruntime.Manifest, bodyexecution.CompositionCases, bodyexecution.CompositionCases) {
	t.Helper()
	read := func(name string) []byte {
		raw, err := os.ReadFile("../../../examples/package-caller-construction/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	cases := func(name string) bodyexecution.CompositionCases {
		suite, err := bodyexecution.DecodeCompositionCases(read(name))
		if err != nil {
			t.Fatal(err)
		}
		return suite
	}
	m := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "app/retry", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "app/retry", Name: "app", Imports: []string{"tools/budget"}, Sources: []packageruntime.Source{
				{Filename: "app.gooo.fixture", Content: string(read("app.gooo.fixture"))}}},
			{Path: "tools/budget", Name: "budgetplan", Sources: []packageruntime.Source{
				{Filename: "budget.gooo.fixture", Content: string(read("budget.gooo.fixture"))}}},
		}}
	return m, cases("construction-cases.json"), cases("evaluation-cases.json")
}

func TestWorkspaceConstructionImportedCallerBudgetReplay(t *testing.T) {
	m, feedback, evaluation := callerConstructionFixture(t)
	baseline, err := ExecuteWorkspace(context.Background(), m, feedback, "", "")
	if err != nil || baseline.Runtime.FinitePassed != 0 || baseline.Runtime.FiniteTotal != 1 {
		t.Fatal(err, baseline.Runtime)
	}
	for _, budget := range []int{5, 6} {
		r, err := ConstructWorkspace(context.Background(), m, feedback, evaluation, ConstructOptions{ProgramBudget: budget})
		if err != nil {
			t.Fatal(err)
		}
		c := r.Construction
		if c.Schema != "gooo/joint-construction/v6" || len(c.Attempts) != budget || c.CandidateSpace != "6" ||
			c.Attempts[1].Rejection == nil || c.Attempts[2].Rejection == nil || c.Attempts[4].Runtime.Outcomes == nil {
			t.Fatal("missing original rejected or faulted attempts", c)
		}
		if c.Attempts[4].Runtime.Outcomes.Faulted != 1 {
			t.Fatal(c.Attempts[4].Runtime.Outcomes)
		}
		wantPassed, wantDecision := 1, "PARTIAL_FINITE"
		if budget == 6 {
			wantPassed, wantDecision = 4, "COMPLETE_FINITE"
		}
		if c.Decision != wantDecision || r.Evaluation.Runtime.FinitePassed != wantPassed || r.Evaluation.Runtime.FiniteTotal != 4 ||
			r.Evaluation.InputSeparation.OtherInputs != 4 || r.Evaluation.ConstructionReplayed {
			t.Fatal(r.Evaluation, c.Decision)
		}
		if r.Program.PureCalls == nil || len(r.Program.PureCalls.Activities) != 2 ||
			!strings.Contains(string(mustConstructionJSON(t, r)), "9007199254740993") {
			t.Fatal("package identity or exact integer lost")
		}
		var saved ConstructionResult
		if err := json.Unmarshal(mustConstructionJSON(t, r), &saved); err != nil {
			t.Fatal(err)
		}
		replay, err := ReplayWorkspaceConstruction(context.Background(), m, saved, evaluation, "")
		if err != nil || replay.ReplayedFromSHA256 == "" || !replay.Evaluation.ConstructionReplayed ||
			replay.Evaluation.NewModelCalls != 0 || replay.Evaluation.Runtime.FinitePassed != wantPassed {
			t.Fatal(err, replay.Evaluation)
		}
	}
}

func mustConstructionJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestWorkspaceConstructionBindsOriginalCasesAndPackageIdentity(t *testing.T) {
	m, feedback, evaluation := callerConstructionFixture(t)
	r, err := ConstructWorkspace(context.Background(), m, feedback, evaluation, ConstructOptions{ProgramBudget: 1})
	if err != nil {
		t.Fatal(err)
	}
	raw := mustConstructionJSON(t, r)
	for name, change := range map[string]func(*ConstructionResult){
		"schema":      func(v *ConstructionResult) { v.Schema = "unknown" },
		"entry":       func(v *ConstructionResult) { v.Program.Entry.Activity = "Other" },
		"joint entry": func(v *ConstructionResult) { v.Construction.Initial.Plan.EntryActivity = "Other" },
		"caller inputs": func(v *ConstructionResult) {
			v.ConstructionCases.Cases[0].Inputs["app/retry:Main"] = json.RawMessage(`{"used":0,"limit":8}`)
		},
		"caller expectation": func(v *ConstructionResult) {
			v.ConstructionCases.Cases[0].Expected["app/retry:Main"] = json.RawMessage(`0`)
		},
		"source identity": func(v *ConstructionResult) { v.Program.PureCalls.Activities[0].ActivityID = "other://activity" },
		"lowered source":  func(v *ConstructionResult) { v.Program.Source += "\n" },
	} {
		t.Run(name, func(t *testing.T) {
			var changed ConstructionResult
			if err := json.Unmarshal(raw, &changed); err != nil {
				t.Fatal(err)
			}
			change(&changed)
			if _, err := ReplayWorkspaceConstruction(context.Background(), m, changed, evaluation, ""); err == nil {
				t.Fatal("changed binding accepted")
			}
		})
	}
	m.Packages[1].Sources[0].Content += "\n"
	if _, err := ReplayWorkspaceConstruction(context.Background(), m, r, evaluation, ""); err == nil {
		t.Fatal("changed package source accepted")
	}
}

func TestWorkspaceConstructionPreflightAndCancellation(t *testing.T) {
	m, feedback, evaluation := callerConstructionFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ConstructWorkspace(ctx, m, feedback, evaluation, ConstructOptions{ProgramBudget: 6}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := ConstructWorkspace(nil, m, feedback, evaluation, ConstructOptions{ProgramBudget: 6}); err == nil {
		t.Fatal("nil context accepted")
	}
	evaluation.Cases[0].Inputs["missing:Main"] = json.RawMessage(`1`)
	_, err := ConstructWorkspace(context.Background(), m, feedback, evaluation, ConstructOptions{ProgramBudget: 6, ModelPath: "missing.json"})
	if err == nil || !strings.Contains(err.Error(), "unknown package activity") {
		t.Fatal("evaluation was not checked before construction", err)
	}
}

func TestWorkspaceConstructionAcrossBoundActivities(t *testing.T) {
	m := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "core", Name: "core", Sources: []packageruntime.Source{{Filename: "core.gooo", Content: `package core
namespace core
entity Integer id "bound://integer"
entity Box id "bound://box" fields { field value id "bound://box/value" type integer required one }
activity Choose(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "double" field_value at "0" alternative "input * 2" intent "Double the value."
 value_case "[0]" -> "{\"value\":0}" attempts "2"
}`}}},
			{Path: "app", Name: "app", Imports: []string{"core"}, Sources: []packageruntime.Source{{Filename: "app.gooo", Content: `package app
namespace app
import core "core"
activity Main(Box) -> Integer computes "return input.value + 1"
bind core.Choose.result -> Main.input
`}}},
		}}
	feedback, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"core:Choose":3},"expected":{"app:Main":7}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := ConstructWorkspace(context.Background(), m, feedback, feedback, ConstructOptions{ProgramBudget: 2})
	if err != nil || r.Construction.Decision != "COMPLETE_FINITE" || len(r.Construction.Attempts) != 2 ||
		r.Program.PureCalls != nil || len(r.Construction.Selected.Plan.Edges) != 1 || r.Evaluation.InputSeparation.ConstructionInputs != 1 {
		t.Fatal(err, r.Construction.Decision)
	}
	replay, err := ReplayWorkspaceConstruction(context.Background(), m, r, feedback, "")
	if err != nil || replay.Evaluation.Runtime.FinitePassed != 1 {
		t.Fatal(err)
	}
}

func TestWorkspaceConstructionEvaluationDoesNotSelect(t *testing.T) {
	m, feedback, evaluation := callerConstructionFixture(t)
	first, err := ConstructWorkspace(context.Background(), m, feedback, evaluation, ConstructOptions{ProgramBudget: 1})
	if err != nil {
		t.Fatal(err)
	}
	evaluation.Cases[0].Expected["app/retry:Main"] = json.RawMessage(`999`)
	changed, err := ConstructWorkspace(context.Background(), m, feedback, evaluation, ConstructOptions{ProgramBudget: 1})
	if err != nil || first.Construction.SelectedSource != changed.Construction.SelectedSource ||
		changed.Construction.SelectedAttempt != first.Construction.SelectedAttempt || changed.Evaluation.Runtime.FinitePassed != 0 {
		t.Fatal("evaluation leaked into selection", err)
	}
	// Source holdouts are still observations. A conflicting holdout cannot choose the body.
	m.Packages[1].Sources[0].Content = strings.Replace(m.Packages[1].Sources[0].Content,
		`holdout_value_case "[{\"used\":9,\"limit\":8}]" -> "{\"next\":8,\"exhausted\":true}"`,
		`holdout_value_case "[{\"used\":9,\"limit\":8}]" -> "{\"next\":999,\"exhausted\":true}"`, 1)
	selected, err := ConstructWorkspace(context.Background(), m, feedback, evaluation, ConstructOptions{ProgramBudget: 6})
	if err != nil || selected.Construction.Decision != "COMPLETE_FINITE" || selected.Construction.SelectedAttempt != 5 ||
		selected.Construction.Attempts[5].FillCandidates[0].HoldoutCasesPassed != 0 {
		t.Fatal("holdout changed construction completion", err)
	}
}

func TestWorkspaceConstructionImportedIntegerSearch(t *testing.T) {
	m := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "core", Name: "core", Sources: []packageruntime.Source{{Filename: "search.gooo", Content: `package core
namespace core
entity Integer id "search://integer"
activity Choose(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__" assembling {
 search hole "value" grammar "integer-offset-constant/v1" intent "Use the input value." max_candidates "8"
 case "0" -> "0" attempts "5"
}`}}},
			{Path: "app", Name: "app", Imports: []string{"core"}, Sources: []packageruntime.Source{{Filename: "app.gooo", Content: `package app
namespace app
import core "core"
activity Main(Integer) -> Integer computes "return core.Choose(input) + input"
`}}},
		}}
	feedback, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"app:Main":3},"expected":{"app:Main":6}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := ConstructWorkspace(context.Background(), m, feedback, feedback, ConstructOptions{ProgramBudget: 5})
	if err != nil || r.Construction.Decision != "COMPLETE_FINITE" || r.Construction.CandidateKinds[0] != "source_search_index" {
		t.Fatal(err, r.Construction.Decision)
	}
	if _, err := ReplayWorkspaceConstruction(context.Background(), m, r, feedback, ""); err != nil {
		t.Fatal(err)
	}
}
