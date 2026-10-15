package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func criteriaChallengeFixture() string {
	source := strings.Replace(criteriaQuestionFixture(), `alternative "0-input"`, `alternative "input+0"`, 1)
	return strings.Replace(source, "observe {", `observe { criteria_strategy "distinguish_then_check"`, 1)
}

func TestCriteriaChallengeFindsSharedWrongAnswerAndRetainsIt(t *testing.T) {
	source := criteriaChallengeFixture()
	r := generateCriteria(t, source)
	a, o := r.Report.RecordAssembly, r.Report.RecordAssembly.CriteriaQuestions
	if !strings.Contains(o.Scope, "only initial and acquired inputs are predicate-scored") {
		t.Fatal("question scope omits initial criteria evaluation", o.Scope)
	}
	if a.Status != "PARTIAL_FINITE" || o.Schema != "gooo/criteria-questions/v2" || o.Status != "NO_SURVIVING_CANDIDATE" ||
		o.ResolvedMask != nil || a.Criteria.Passed != 1 || a.Criteria.Failed != 1 || a.Criteria.Total != 2 || len(o.AcquiredInputs) != 1 ||
		o.Rounds[0].RecommendationReason != "CHECK_CRITERIA" || o.Rounds[0].Probes[0].SeparatedPairs != 0 {
		t.Fatal("shared answer escaped source criteria", a)
	}
	for _, check := range o.Evaluations {
		if check.Result.InputJSON == "[-5]" {
			t.Fatal("unchosen question pre-scored")
		}
	}
	raw, _ := json.Marshal(r)
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "shared.gooo", []byte(source), saved); err != nil {
		t.Fatal(err)
	}
	candidate, _, err := RealizeObservedRecordCandidate(context.Background(), "shared.gooo", []byte(source), "Select", a, 1)
	if err != nil || candidate.Attempt.Criteria.Failed != 1 || candidate.Attempt.Criteria.Total != 2 {
		t.Fatal("acquired shared check disappeared", err, candidate)
	}
	saved.Report.RecordAssembly.CriteriaQuestions.Rounds[0].RecommendationReason = "SEPARATE_OUTPUTS"
	if _, err := RealizeSourceAssembly(context.Background(), "shared.gooo", []byte(source), saved); err == nil {
		t.Fatal("changed acquisition reason replayed")
	}
}

func TestCriteriaChallengePrefersDisagreementThenChecksSoleSurvivor(t *testing.T) {
	source := strings.Replace(criteriaQuestionFixture(), "observe {", `observe { criteria_strategy "distinguish_then_check"`, 1)
	r := generateCriteria(t, source).Report.RecordAssembly
	o := r.CriteriaQuestions
	if o.Status != "ROUND_BUDGET" || o.ResolvedMask != nil || o.Rounds[0].RecommendationReason != "SEPARATE_OUTPUTS" ||
		o.Rounds[1].RecommendationReason != "CHECK_CRITERIA" || o.Rounds[1].RecommendedIndex == nil {
		t.Fatal(o)
	}
	source = strings.Replace(source, `max_rounds "1"`, `max_rounds "2"`, 1)
	r = generateCriteria(t, source).Report.RecordAssembly
	o = r.CriteriaQuestions
	if o.Status != "ONE_SURVIVING_CANDIDATE" || o.ResolvedMask == nil || *o.ResolvedMask != 1 || r.Criteria.Total != 3 ||
		len(o.AcquiredInputs) != 2 || o.AcquiredInputs[1] != "[-5]" {
		t.Fatal(o)
	}
}

func TestCriteriaChallengeUnavailableQuestionDoesNotProveAgreement(t *testing.T) {
	source := strings.Replace(criteriaChallengeFixture(), "value:input", "value:input/(1-input)", 1)
	source = strings.Replace(source, `alternative "input+0"`, `alternative "input/(1-input)+0"`, 1)
	source = strings.Replace(source, `criteria_input "[7]"`, `criteria_input "[1]"`, 1)
	source = strings.Replace(source, `criteria_input "[-5]"`, "", 1)
	source = strings.Replace(source, "return input1.value == 0-input0", "return true", 1)
	o := generateCriteria(t, source).Report.RecordAssembly.CriteriaQuestions
	if o.Status != "QUESTION_OUTPUTS_UNAVAILABLE" || o.ResolvedMask != nil || len(o.AcquiredInputs) != 0 ||
		o.Rounds[0].Probes[0].OutputEvidence.Unavailable != 2 || o.Rounds[0].Probes[0].OutputEvidence.Available != 0 {
		t.Fatal(o)
	}
}
