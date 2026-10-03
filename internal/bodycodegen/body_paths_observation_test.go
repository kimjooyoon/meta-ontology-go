package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func observationFixture(t *testing.T) ([]byte, pathplan.Document, TypedPathOptions) {
	t.Helper()
	source, document := diagnosisFixture(t)
	source = append(source, []byte("activity Expected(Integer) -> Integer computes \"return (2 - input)\"\n")...)
	document.MaxAttempts = 2
	document.Plan.Decisions[0].Intent = "2에서 입력을 뺀다. Subtract the input from two."
	return source, document, TypedPathOptions{Observation: &PathObservationOptions{
		Inputs: []int64{2, 3, 0}, MaxCandidates: 2, MaxRounds: 2, OracleActivity: "Expected"}}
}

func TestPathObservationSeparatesTwoConditionalDecisionsAcrossRounds(t *testing.T) {
	plan := pathplan.Plan{Schema: pathplan.Schema, Base: bodyplan.Plan{Schema: bodyplan.Schema, ID: "two-branches",
		Name: "Probe", ResultType: decision.TypeInt,
		Expressions: []bodyplan.Expr{{Kind: "input", Name: "input"}, {Kind: "int", Int: 0},
			{Kind: "binary", Operation: "less_than", Left: 0, Right: 1},
			{Kind: "binary", Operation: "subtract", Left: 0, Right: 1},
			{Kind: "binary", Operation: "subtract", Left: 0, Right: 1}},
		Statements: []bodyplan.Stmt{{Kind: "return", Expr: 3}, {Kind: "return", Expr: 4},
			{Kind: "if", Expr: 2, Then: []int{0}, Else: []int{1}}}, Root: []int{2}}}
	for i := range 2 {
		plan.Decisions = append(plan.Decisions, pathplan.Choice{ID: fmt.Sprintf("order_%d", i),
			Kind: pathplan.OperandOrder, Target: 3 + i, Intent: "Choose the operand order.", Fallback: "layout_forward",
			Options: []pathplan.Option{{Label: "layout_forward"}, {Label: "layout_reverse", Reverse: true}}})
	}
	prepared, err := pathplan.Prepare(plan)
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(prepared.Fallback().GoooSource() + "\nactivity Expected(Integer) -> Integer computes \"if input < 0 { return (0 - input) } else { return input }\"\n")
	doc := pathplan.Document{Schema: pathplan.DocumentSchema, Plan: plan, MaxAttempts: 4, TestCases: []pathplan.TestCase{{Input: 0, Expected: 0}}}
	for _, mode := range []struct {
		budget  int
		reuse   bool
		resolve bool
	}{{1, false, false}, {2, false, false}, {1, true, false}, {2, true, false}, {1, true, true}, {2, true, true}} {
		budget := mode.budget
		result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "", TypedPathOptions{
			Observation: &PathObservationOptions{Inputs: []int64{-1, 1}, MaxCandidates: 4, MaxRounds: budget, OracleActivity: "Expected", ReuseProbeOutputs: mode.reuse, ResolveUniqueCandidate: mode.resolve}})
		if err != nil {
			t.Fatal(err)
		}
		r := result.Report.BodyPaths.Observation
		if len(r.Rounds) != budget+1 || len(r.Rounds[0].Ranking.SurvivingMasks) != 4 || len(r.Rounds[1].Ranking.SurvivingMasks) != 2 {
			t.Fatal(r)
		}
		if budget == 1 && r.Status != "ROUND_BUDGET" {
			t.Fatal(r)
		}
		if budget == 2 && (r.Status != "ONE_SURVIVING_CANDIDATE" || r.Rounds[2].Ranking.SurvivingMasks[0] != 1) {
			t.Fatal(r)
		}
		if mode.resolve && pathResolved(result.Report.BodyPaths) != (budget == 2) {
			t.Fatal("resolved before both branches were distinguished")
		}
		if _, _, bad := observedPathContract(result.Report.BodyPaths); bad {
			t.Fatal("invalid multi-round binding")
		}
		if mode.reuse {
			last := r.Rounds[budget]
			if last.Ranking.EvaluationAttempts != 0 || last.Reuse.TotalEvaluationAttempts != 12 ||
				last.Reuse.TotalCachedComparisons != 2+2*budget {
				t.Fatal("cached rounds repeated evaluation", r)
			}
		}
		if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, doc, result); err != nil {
			t.Fatal("multi-round observation failed replay", err)
		}
	}
}

func TestPathObservationComposesWithThreeChoiceModelFeedback(t *testing.T) {
	source, doc := threeNativeFixture(t)
	source = append(source, []byte("activity Expected(Integer) -> Integer computes \"return (7 - input)\"\n")...)
	doc.TestCases = []pathplan.TestCase{{Input: 10, Expected: -3}}
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "ThreeTest", doc, writeThreeContractModel(t),
		TypedPathOptions{StepAttempts: 1, FeedbackRounds: 7, Observation: &PathObservationOptions{
			Inputs: []int64{0, 3, -1}, MaxCandidates: 8, MaxRounds: 2, OracleActivity: "Expected"}})
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p.Observation.EffectiveCases != 2 || p.Search.SelectedTrainingPassed != 2 || p.Search.Selection.ModelCalls < 1 ||
		p.ModelContext == nil || p.Observation.Status != "ONE_SURVIVING_CANDIDATE" || p.DeclaredTestCases != 1 {
		t.Fatal("observation/model feedback did not compose", p)
	}
	if _, _, bad := observedPathContract(p); bad {
		t.Fatal("model context changed original observation identity")
	}
}

func TestPathObservationAppendsOracleEvidenceAndEmitsResolvedBody(t *testing.T) {
	source, document, options := observationFixture(t)
	before, _ := json.Marshal(document)
	sourceSHA := digest(source)
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", document, "", options)
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	r := p.Observation
	if r == nil || r.Status != "ONE_SURVIVING_CANDIDATE" || len(r.Rounds) != 2 ||
		r.OracleEvaluations != 2 || r.EffectiveCases != 2 || r.Rounds[0].Observation.Input != 3 ||
		r.Rounds[0].Observation.Expected != -1 || len(r.Rounds[0].Ranking.SurvivingMasks) != 2 ||
		len(r.Rounds[1].Ranking.SurvivingMasks) != 1 || r.Rounds[1].Ranking.SurvivingMasks[0] != 1 ||
		p.DeclaredTestCases != 1 || p.Search.TrainingTotal != 2 || p.Search.SelectedTrainingPassed != 2 ||
		p.Search.Selection.ModelCalls != 0 || result.Report.RepositoryWrites != 0 ||
		!strings.Contains(result.Source, "return (2 - input)") || p.Timing.ObservationMS <= 0 {
		t.Fatalf("observation loop: %+v", p)
	}
	if _, _, bad := observedPathContract(p); bad {
		t.Fatal("invalid derived suite")
	}
	if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, document, result); err != nil {
		t.Fatal("source-bound execution replay lost new observations", err)
	}
	for _, id := range []string{"typed_path_finite_accuracy", "typed_path_observation_binding"} {
		if d := completenessDimensionByID(t, result.Report.CompletenessReceipt, id); d.Status != "PASS" {
			t.Fatal(d)
		}
	}
	after, _ := json.Marshal(document)
	if string(before) != string(after) || digest(source) != sourceSHA {
		t.Fatal("caller inputs changed")
	}
	g, _ := NewTypedPathGenerator("")
	retained, err := g.Generate(context.Background(), "fixture.gooo", source, "Probe", document, options)
	if err != nil || retained.Source != result.Source || retained.Report.BodyPaths.Observation.EffectiveCasesSHA != r.EffectiveCasesSHA {
		t.Fatal("retained generator lost observation options", err)
	}
}

func TestPathObservationNoOracleAndFiniteBudgetsRemainExplicit(t *testing.T) {
	for _, variant := range []string{"missing", "candidate", "cases", "rounds", "agreement"} {
		t.Run(variant, func(t *testing.T) {
			source, doc, options := observationFixture(t)
			want := "ORACLE_UNAVAILABLE"
			switch variant {
			case "missing":
				options.Observation.OracleActivity = ""
			case "candidate":
				options.Observation.MaxCandidates = 1
				want = "CANDIDATE_BUDGET"
			case "cases":
				for len(doc.TestCases) < 128 {
					doc.TestCases = append(doc.TestCases, doc.TestCases[0])
				}
				want = "CASE_BUDGET"
			case "rounds":
				options.Observation.MaxRounds = 1
				want = "ONE_SURVIVING_CANDIDATE"
			case "agreement":
				options.Observation.Inputs = []int64{2, 2}
				want = "NO_DISTINGUISHING_INPUT"
			}
			result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "", options)
			if err != nil || result.Report.BodyPaths.Observation.Status != want {
				t.Fatal(variant, want, err, result.Report.BodyPaths)
			}
			if _, _, bad := observedPathContract(result.Report.BodyPaths); bad {
				t.Fatal("invalid finite record")
			}
			if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, doc, result); err != nil {
				t.Fatal("bounded observation did not replay", err)
			}
		})
	}
}

func TestPathObservationReplayRecomputesCandidateOutputs(t *testing.T) {
	source, doc, options := observationFixture(t)
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "", options)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := doc.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	result.Report.BodyPaths.Observation.Rounds[0].Ranking.Probes[1].Outputs[0]++
	if _, err := replayPathObservation(context.Background(), "fixture.gooo", source, "Probe", prepared, doc.TestCases, result.Report.BodyPaths); err == nil {
		t.Fatal("changed candidate actual accepted by independent replay")
	}
}

func TestPathObservationRejectsConflictingAndInvalidOracleBeforeModelLoad(t *testing.T) {
	for _, variant := range []string{"same", "unknown", "contradiction", "type", "budget", "identifier"} {
		t.Run(variant, func(t *testing.T) {
			source, document, options := observationFixture(t)
			switch variant {
			case "same":
				options.Observation.OracleActivity = "Probe"
			case "unknown":
				options.Observation.OracleActivity = "Missing"
			case "contradiction":
				source = []byte(strings.ReplaceAll(string(source), "return (2 - input)", "return (3 - input)"))
			case "type":
				source = append(source, []byte("entity Boolean id \"path-diagnosis://boolean\"\nactivity BoolOracle(Integer) -> Boolean computes \"return true\"\n")...)
				options.Observation.OracleActivity = "BoolOracle"
			case "budget":
				options.Observation.MaxRounds = 9
			case "identifier":
				options.Observation.OracleActivity = "bad/name"
			}
			_, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", document, "missing-model.json", options)
			if err == nil || strings.Contains(err.Error(), "load explicit structural model") {
				t.Fatal("oracle error reached model load", err)
			}
		})
	}
	source, document, options := observationFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GenerateWithTypedPathOptions(ctx, "fixture.gooo", source, "Probe", document, "", options); err == nil {
		t.Fatal("cancellation ignored")
	}
}

func TestPathObservationReceiptRejectsChangedEvidence(t *testing.T) {
	source, document, options := observationFixture(t)
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", document, "", options)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result.Report.BodyPaths)
	for _, mutate := range []func(*BodyPathReceipt){
		func(p *BodyPathReceipt) { p.Observation.EffectiveCases++ },
		func(p *BodyPathReceipt) { p.Observation.InitialCases[0].Expected++ },
		func(p *BodyPathReceipt) { p.Observation.Options.MaxRounds++ },
		func(p *BodyPathReceipt) { p.Observation.SourceSHA256 = digest([]byte("different")) },
		func(p *BodyPathReceipt) { p.Observation.Rounds[0].Observation.Input++ },
		func(p *BodyPathReceipt) { p.Observation.Rounds[0].Ranking.CasesSHA256 = "changed" },
		func(p *BodyPathReceipt) { p.Observation.OracleInitialCases[0].Actual++ },
		func(p *BodyPathReceipt) { p.Observation.OracleEvaluations++ },
	} {
		var p BodyPathReceipt
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatal(err)
		}
		mutate(&p)
		if _, _, bad := observedPathContract(&p); !bad {
			t.Fatal("changed observation accepted")
		}
	}
}
