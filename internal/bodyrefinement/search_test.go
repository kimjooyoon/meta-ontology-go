package bodyrefinement

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func searchFeedbackFixture(t *testing.T) ([]byte, bodyexecution.CompositionCases, Options) {
	t.Helper()
	_, _, options := refinementFixture(t)
	read := func(name string) []byte {
		raw, err := os.ReadFile("../../examples/search-feedback/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	suite, err := bodyexecution.DecodeCompositionCases(read("feedback-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := bodyexecution.DecodeCompositionCases(read("evaluation-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	options.Activity, options.SearchPolicy, options.MaxRounds = "Add", true, 6
	options.PolicySource, options.Evaluation = read("policy.gooo.fixture"), &evaluation
	return read("source.gooo.fixture"), suite, options
}

func TestSearchPolicyExpandsDeclaredSpaceAndChangesGrammar(t *testing.T) {
	source, suite, options := searchFeedbackFixture(t)
	result, err := Run(context.Background(), "search.gooo", source, suite, options)
	if err != nil || result.Status != "PASS" || result.EvaluationStatus != "PASS" || len(result.Rounds) != 4 || !result.SearchPolicy {
		t.Fatal("source-declared transitions did not finish the finite example", err, result.Status, len(result.Rounds))
	}
	for i, action := range []string{"INCORPORATE", "ADVANCE_SEARCH", "ADVANCE_SEARCH", "STOP"} {
		round := result.Rounds[i]
		if round.Decision.Action != action || round.Search == nil || round.PolicyRuntime.ModelCalls != 0 {
			t.Fatal("transition was not decided by Gooo", i, round.Decision)
		}
		if err := bodyexecution.VerifyComposition(context.Background(), "search.gooo", []byte(round.Source), round.Composition); err != nil {
			t.Fatal("a retained round cannot replay", i, err)
		}
	}
	if result.Rounds[1].Search.Omitted != 4 || result.Rounds[2].Search.Omitted != 0 || result.Rounds[3].Search.Grammar != "integer-hole-residual/v1" || result.Rounds[3].Runtime.FinitePassed != 2 {
		t.Fatal("candidate cap and grammar transitions were not observed")
	}
	selected := result.Rounds[result.SelectedRound]
	replay, err := bodyexecution.ExecuteComposition(context.Background(), "search.gooo", []byte(selected.Source), selected.Composition, suite, options.GoBinary)
	if err != nil || replay.ModelCalls != 0 || replay.FinitePassed != 2 {
		t.Fatal("selected grammar did not replay without inference", err)
	}
}

func TestSearchPolicyKeepsUnavailableAlternativesAndRoundCapsPartial(t *testing.T) {
	for _, mode := range []string{"none", "round-cap", "return-to-initial"} {
		t.Run(mode, func(t *testing.T) {
			source, suite, options := searchFeedbackFixture(t)
			if mode == "round-cap" {
				options.MaxRounds = 2
			} else {
				var lines []string
				for line := range strings.SplitSeq(string(source), "\n") {
					if strings.Contains(line, "search_alternative") {
						continue
					}
					lines = append(lines, line)
				}
				source = []byte(strings.Join(lines, "\n"))
				if mode == "return-to-initial" {
					source = []byte(strings.Replace(string(source), `case "2"`, `search_alternative "again" grammar "integer-offset-constant/v1" max_candidates "2"`+"\n    case \"2\"", 1))
				}
			}
			result, err := Run(context.Background(), "search.gooo", source, suite, options)
			if err != nil || result.Status != "PROGRESS" || len(result.Rounds) != 2 || result.Rounds[1].Decision.Action != "STOP" {
				t.Fatal("partial state or visited search boundary lost", err, result.Status, len(result.Rounds))
			}
		})
	}
}

func TestSearchPolicyEvaluationDoesNotChooseGrammar(t *testing.T) {
	source, suite, options := searchFeedbackFixture(t)
	first, err := Run(context.Background(), "search.gooo", source, suite, options)
	if err != nil {
		t.Fatal(err)
	}
	options.Evaluation.Cases[0].Expected["Add"] = json.RawMessage(`999`)
	second, err := Run(context.Background(), "search.gooo", source, suite, options)
	if err != nil || second.Status != "PROGRESS" || second.FeedbackStatus != "PASS" || len(second.Rounds) != len(first.Rounds) {
		t.Fatal("changed evaluation did not stay separate", err)
	}
	for i, round := range first.Rounds {
		if !reflect.DeepEqual(round.Decision, second.Rounds[i].Decision) || round.Source != second.Rounds[i].Source {
			t.Fatal("final evaluation changed adaptive source decisions", i)
		}
	}
}

func TestSearchPolicyCannotSelectAnUndeclaredAlternative(t *testing.T) {
	source, suite, options := searchFeedbackFixture(t)
	options.PolicySource = []byte(strings.Replace(string(options.PolicySource), "next_search_id: input.next_search_id", `next_search_id: "undeclared"`, 1))
	result, err := Run(context.Background(), "search.gooo", source, suite, options)
	if err == nil || !strings.Contains(err.Error(), "next unvisited source alternative") || result.Status != "FAIL_CLOSED" || len(result.Rounds) != 2 || result.Rounds[1].Decision.NextSearchID != "undeclared" {
		t.Fatal("undeclared alternative executed or decision evidence lost", err, result.Status, len(result.Rounds))
	}
}
