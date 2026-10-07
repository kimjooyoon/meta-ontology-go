package bodyrefinement

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func refinementFixture(t *testing.T) ([]byte, bodyexecution.CompositionCases, Options) {
	t.Helper()
	read := func(name string) []byte {
		raw, err := os.ReadFile(filepath.Join("..", "..", "examples", name))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	source := read("body-codegen/record-candidate-continuation.gooo.fixture")
	source, err := bodycodegen.ReviseAssemblyBudget(context.Background(), "target.gooo", source, "Select", 2)
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionCases(read("body-codegen/record-field-updates-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	return source, suite, Options{Activity: "Select", MaxAttempts: 8, MaxRounds: 4,
		PolicySource: read("assembly-feedback/policy.gooo.fixture"), GoBinary: filepath.Join(runtime.GOROOT(), "bin", "go")}
}

func TestGoooFeedbackRefinesAndReplaysRetainedProgram(t *testing.T) {
	source, suite, options := refinementFixture(t)
	result, err := Run(context.Background(), "target.gooo", source, suite, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rounds) != 3 || result.SelectedRound != 2 || result.Status != "PASS" || result.StopReason != "POLICY_STOP" || result.EvaluationStatus != "UNKNOWN" {
		t.Fatal("feedback did not reach its declared boundary", result.Status, len(result.Rounds), result.SelectedRound)
	}
	for i, round := range result.Rounds {
		if round.Observation.Attempts != 2<<i || round.Runtime.FiniteTotal != 14 || !round.Runtime.RuntimeReplayed ||
			!round.PolicyRuntime.RuntimeReplayed || round.PolicyRuntime.FiniteTotal != 0 || round.PolicyRuntime.FinitePassed != 0 {
			t.Fatal("round lost actual observations", i)
		}
		if err := bodyexecution.VerifyComposition(context.Background(), "target.gooo", []byte(round.Source), round.Composition); err != nil {
			t.Fatal("saved round cannot replay", i, err)
		}
	}
	if result.Rounds[0].Runtime.FinitePassed >= 14 || result.Rounds[2].Runtime.FinitePassed != 14 {
		t.Fatal("example did not improve its finite feedback result")
	}
	selected := result.Rounds[result.SelectedRound]
	again, err := bodyexecution.ExecuteComposition(context.Background(), "target.gooo", []byte(selected.Source), selected.Composition, suite, options.GoBinary)
	if err != nil || again.FinitePassed != 14 || again.ModelCalls != 0 {
		t.Fatal("retained program failed without new model inference", err)
	}
}

func TestGoooFeedbackEvaluationCannotChangePolicyDecisions(t *testing.T) {
	source, suite, options := refinementFixture(t)
	var err error
	source, err = bodycodegen.ReviseAssemblyBudget(context.Background(), "target.gooo", source, "Select", 8)
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := Run(context.Background(), "target.gooo", source, suite, options)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(suite)
	evaluation, _ := bodyexecution.DecodeCompositionCases(raw)
	evaluation.Cases[0].Expected["Label"] = json.RawMessage(`"a deliberately different final obligation"`)
	options.Evaluation = &evaluation
	withEvaluation, err := Run(context.Background(), "target.gooo", source, suite, options)
	if err != nil || withEvaluation.Evaluation == nil || withEvaluation.EvaluationStatus != "PROGRESS" {
		t.Fatal("final evaluation was not retained", err)
	}
	if withEvaluation.Rounds[0].Decision != baseline.Rounds[0].Decision || withEvaluation.Rounds[0].Composition.GeneratedSHA256 != baseline.Rounds[0].Composition.GeneratedSHA256 {
		t.Fatal("evaluation affected an adaptive choice")
	}
	if withEvaluation.Status != "PROGRESS" || withEvaluation.FeedbackStatus != "PASS" || withEvaluation.SelectedRound != 0 {
		t.Fatal("a partial result became a complete claim")
	}
}

func TestGoooFeedbackUsesChangedPolicyAndStopsAtRoundLimit(t *testing.T) {
	source, suite, options := refinementFixture(t)
	policy := strings.Replace(string(options.PolicySource), "input.attempts * 2", "input.attempts + 1", 1)
	policy = strings.Replace(policy, "input.round >= input.round_limit", "input.round > input.round_limit", 1)
	policy = strings.ReplaceAll(policy, "\"INCORPORATE\"", "\"CONTINUE\"")
	options.PolicySource, options.MaxRounds = []byte(policy), 2
	result, err := Run(context.Background(), "target.gooo", source, suite, options)
	if err != nil || len(result.Rounds) != 2 || result.Rounds[1].Observation.Attempts != 3 || result.StopReason != "ROUND_LIMIT" {
		t.Fatal("Gooo policy or round bound was not applied", err, result.StopReason)
	}
	if result.Status != "PROGRESS" || result.Rounds[1].Decision.Action != "CONTINUE" {
		t.Fatal("round-limit stop lost the unconsumed policy decision")
	}
}

func TestGoooFeedbackRejectsNonadvancingPolicyAndPreservesEvidence(t *testing.T) {
	source, suite, options := refinementFixture(t)
	policy := strings.Replace(string(options.PolicySource), "let next = input.attempts * 2", "let next = input.attempts", 1)
	options.PolicySource = []byte(strings.ReplaceAll(policy, "\"INCORPORATE\"", "\"CONTINUE\""))
	result, err := Run(context.Background(), "target.gooo", source, suite, options)
	if err == nil || result.Status != "FAIL_CLOSED" || len(result.Rounds) != 1 || !result.Rounds[0].PolicyRuntime.RuntimeReplayed {
		t.Fatal("nonadvancing policy was repeated or its observation lost", err)
	}
}

func TestGoooFeedbackRejectsInvalidBoundsBeforeModel(t *testing.T) {
	source, suite, options := refinementFixture(t)
	options.ModelPath = "/missing/model.json"
	options.MaxRounds = 9
	result, err := Run(context.Background(), "target.gooo", source, suite, options)
	if err == nil || len(result.Rounds) != 0 || strings.Contains(err.Error(), "model.json") {
		t.Fatal("invalid bounds loaded a model", err)
	}
	options.MaxRounds = 4
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Run(ctx, "target.gooo", source, suite, options); err == nil {
		t.Fatal("canceled refinement succeeded")
	}
}

func TestGoooFeedbackRefinesSourceIRSearch(t *testing.T) {
	_, _, options := refinementFixture(t)
	options.Activity = "Add"
	source := []byte(`package offset
namespace offset
entity Integer id "offset://integer"
activity Add(Integer) -> Integer computes "return __GOOO_BODY_HOLE_offset__" assembling {
    search hole "offset" grammar "integer-offset-constant/v1" intent "Choose a typed offset." max_candidates "16"
    case "2" -> "3"
    attempts "1"
}`)
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Add":2},"expected":{"Add":3}}, {"inputs":{"Add":10},"expected":{"Add":11}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := Run(context.Background(), "offset.gooo", source, suite, options)
	if err != nil || result.Status != "PASS" {
		t.Fatal("source IR search did not reach finite expectations", err, result.Status)
	}
	if len(result.Rounds) < 2 || result.Rounds[0].Composition.Steps[0].Generation.Report.BodySearch == nil {
		t.Fatal("example did not consume source IR search feedback")
	}
	added := 0
	for _, round := range result.Rounds {
		added += round.FeedbackAdded
	}
	if added < 1 {
		t.Fatal("the loop did not promote an actual counterexample")
	}
}

func TestGoooFeedbackMissingGrammarExpressionRemainsIncomplete(t *testing.T) {
	_, _, options := refinementFixture(t)
	options.Activity = "Add"
	source := []byte(`package offset
namespace offset
entity Integer id "offset://integer"
activity Add(Integer) -> Integer computes "return input + __GOOO_BODY_HOLE_offset__" assembling {
    search hole "offset" grammar "integer-offset-constant/v1" intent "Choose a typed offset." max_candidates "16"
    case "2" -> "3"
    attempts "1"
}`)
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Add":2},"expected":{"Add":3}}, {"inputs":{"Add":10},"expected":{"Add":11}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := Run(context.Background(), "offset.gooo", source, suite, options)
	if err != nil || result.Status != "PROGRESS" || len(result.Rounds) != 4 {
		t.Fatal("a missing expression was counted as complete", err, result.Status)
	}
	selected := result.Rounds[result.SelectedRound]
	if selected.Runtime.FinitePassed != 0 || selected.Runtime.FiniteTotal != 2 {
		t.Fatal("unmet nested-hole obligations were hidden")
	}
}
