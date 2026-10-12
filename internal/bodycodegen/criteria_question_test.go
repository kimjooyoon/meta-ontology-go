package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func criteriaQuestionFixture() string {
	s := strings.Replace(typedCriteriaSource, `input "[7]"`, `input "[0]"`, 1)
	return strings.Replace(s, `attempts "2"`, `observe {
 criteria_input "[7]"
 criteria_input "[-5]"
 max_candidates "2"
 max_rounds "1"
 reuse_outputs "true"
 resolve_unique "true"
}
attempts "2"`, 1)
}

func TestCriteriaQuestionsAcquireInputWithoutExpectedRecord(t *testing.T) {
	source := criteriaQuestionFixture()
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "questions.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	o := r.CriteriaQuestions
	if o == nil || o.Status != "ONE_SURVIVING_CANDIDATE" || len(o.InitialInputs) != 1 || len(o.AcquiredInputs) != 1 || o.AcquiredInputs[0] != "[7]" || o.EffectiveInputs != 2 ||
		o.ResolvedMask == nil || *o.ResolvedMask != 1 || r.SelectedMask != 1 || r.ModelCalls != 0 || r.Total != 0 || r.Criteria.Passed != 2 || r.Criteria.Total != 2 || len(r.Attempts) != 1 {
		t.Fatalf("criteria acquisition differs: %+v", r)
	}
	if o.CandidateEvaluations != 6 || o.PredicateEvaluations != 4 || o.ReusedOutputs == 0 || o.ReusedCriteriaResults == 0 {
		t.Fatal("work counts do not describe actual evaluations", o)
	}
	for _, e := range o.Evaluations {
		if e.Result.InputJSON == "[-5]" {
			t.Fatal("unselected question was predicate-scored")
		}
	}
	raw, _ := json.MarshalIndent(result, "", "  ")
	var saved Result
	json.Unmarshal(raw, &saved)
	if _, err := RealizeSourceAssembly(context.Background(), "questions.gooo", []byte(source), saved); err != nil {
		t.Fatal("question history did not replay", err)
	}
	saved.Report.RecordAssembly.CriteriaQuestions.AcquiredInputs[0] = "[-5]"
	if _, err := RealizeSourceAssembly(context.Background(), "questions.gooo", []byte(source), saved); err == nil {
		t.Fatal("tampered acquired question replayed")
	}
}

func TestCriteriaQuestionsUnseenMaskIsNotUnique(t *testing.T) {
	source := strings.Replace(criteriaQuestionFixture(), `max_candidates "2"`, `max_candidates "1"`, 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "questions.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	o := result.Report.RecordAssembly.CriteriaQuestions
	if o.Status != "CANDIDATE_BUDGET" || o.ResolvedMask != nil || len(o.AcquiredInputs) != 0 || o.Rounds[0].Unobserved != 1 {
		t.Fatal(o)
	}
}

func TestCriteriaQuestionsUnknownPredicateDoesNotExcludeCandidate(t *testing.T) {
	source := strings.Replace(criteriaQuestionFixture(), "return input1.value == 0-input0", "if input1.value > 0 { return input0 / (input0-input0) > 0 }; return input1.value == 0-input0", 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "questions.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	o := result.Report.RecordAssembly.CriteriaQuestions
	if o.ResolvedMask != nil || o.Status != "ROUND_BUDGET" || len(o.Rounds[1].UnavailableMasks) != 1 || len(o.Rounds[1].SurvivingMasks) != 1 {
		t.Fatal("unknown candidate vanished", o)
	}
}

func TestCriteriaQuestionsRejectUnknownInputShapeBeforeCandidateExecution(t *testing.T) {
	source := strings.Replace(criteriaQuestionFixture(), `criteria_input "[7]"`, `criteria_input "[true]"`, 1)
	if err := ValidateSourceAssembly(context.Background(), "questions.gooo", []byte(source), "Select"); err == nil {
		t.Fatal("incorrectly typed question accepted")
	}
}
