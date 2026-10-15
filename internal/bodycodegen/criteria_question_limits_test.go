package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestCriteriaQuestionsEqualOutputsLeaveUnaskedPredicatesUnresolved(t *testing.T) {
	source := strings.Replace(criteriaQuestionFixture(), `alternative "0-input"`, `alternative "input+0"`, 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "equal.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	o := result.Report.RecordAssembly.CriteriaQuestions
	if o.Status != "NO_DISTINGUISHING_INPUT" || o.ResolvedMask != nil || len(o.AcquiredInputs) != 0 ||
		len(o.Rounds[0].SurvivingMasks) != 2 || o.PredicateEvaluations != 2 {
		t.Fatal(o)
	}
	for _, e := range o.Evaluations {
		if e.Result.InputJSON != "[0]" {
			t.Fatal("indistinguishable query was silently judged", e)
		}
	}
}

func TestCriteriaQuestionsReuseChangesWorkButNotSelectedMeaning(t *testing.T) {
	source := strings.Replace(criteriaQuestionFixture(), `reuse_outputs "true"`, `reuse_outputs "false"`, 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "fresh.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	o := result.Report.RecordAssembly.CriteriaQuestions
	if result.Report.RecordAssembly.SelectedMask != 1 || o.Status != "ONE_SURVIVING_CANDIDATE" ||
		o.ReusedOutputs != 0 || o.ReusedCriteriaResults != 0 || o.PredicateEvaluations != 6 || o.CandidateEvaluations <= 6 {
		t.Fatal(o)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "fresh.gooo", []byte(source), result); err != nil {
		t.Fatal(err)
	}
}
