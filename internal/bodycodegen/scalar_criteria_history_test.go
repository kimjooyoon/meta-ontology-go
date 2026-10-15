package bodycodegen

import (
	"context"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"strings"
	"testing"
)

func TestScalarCriteriaResumeRetainsOriginalQuestionsAndNewPredicate(t *testing.T) {
	prior := generateScalarCriteria(t, scalarCriteriaSource)
	changed := strings.Replace(scalarCriteriaSource, "return input1 == input0 - 11", "return input1 == 11-input0", 1)
	result, err := ResumeRecordAssembly(context.Background(), "scalar.gooo", []byte(scalarCriteriaSource), []byte(changed), prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	a := result.Report.RecordAssembly
	if a.SelectedMask != 0 || a.Criteria.Passed != 2 || a.Continuation.RecheckedAttempts != 2 || a.Continuation.AddedAttempts != 0 ||
		!equalLocalCallEvidence(a.CriteriaQuestions, prior.Report.RecordAssembly.CriteriaQuestions) {
		t.Fatal(a)
	}
	if a.Criteria.Bindings[0].GeneratedSHA256 == prior.Report.RecordAssembly.Criteria.Bindings[0].GeneratedSHA256 {
		t.Fatal("predicate identity did not change")
	}
	if _, err := RealizeSourceAssembly(context.Background(), "scalar.gooo", []byte(changed), result); err != nil {
		t.Fatal(err)
	}
}
func TestScalarCriteriaDeclinesRecordModelFeatures(t *testing.T) {
	ctx := context.Background()
	p, err := prepareRecordAssembly(ctx, "scalar.gooo", []byte(scalarCriteriaSource), "Fee")
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	// Synthetic loaded-state metadata exercises the decline branch, without a model artifact or prediction.
	g.info.Loaded = true
	r := newRecordAssemblyReceipt([]byte(scalarCriteriaSource), p)
	if err := g.rankRecordAssembly(p, r); err != nil {
		t.Fatal(err)
	}
	if r.ModelCalls != 0 || r.Prediction != nil || r.Context == nil || r.Context.Reason != "SCALAR_CRITERIA_FEATURE_CONTRACT_UNAVAILABLE" {
		t.Fatal(r)
	}
	if err := verifyRecordAssemblyRanking(r, newRecordAssemblyReceipt([]byte(scalarCriteriaSource), p), p); err != nil {
		t.Fatal(err)
	}
}

func TestScalarCriteriaContextExportKeepsRecordFeaturesUnavailable(t *testing.T) {
	r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "scalar.gooo", []byte(scalarCriteriaSource), "Fee", true, jointdecision.RecordGraphSharedFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	if r.Schema != "gooo/scalar-criteria-input-export/v1" || r.Context.Status != "DECLINED_TO_DETERMINISTIC" || r.ValueFlow != nil || r.ModelPredictions != 0 || r.CandidateTests != 0 || r.ExpandedPlan == nil || r.ExpandedPlan.Criteria == nil {
		t.Fatal(r)
	}
}

func TestScalarCriteriaBudgetsAndReceiptDomainRemainExplicit(t *testing.T) {
	source := strings.Replace(scalarCriteriaSource, `max_candidates "2"`, `max_candidates "1"`, 1)
	source = strings.Replace(source, `attempts "2"`, `attempts "1"`, 1)
	result := generateScalarCriteria(t, source)
	a := result.Report.RecordAssembly
	explanation := ExplainCriteriaRejections(a)
	if a.Status != "PARTIAL_FINITE" || len(a.Attempts) != 1 || explanation.Unobserved != 1 ||
		explanation.Status != "EXAMINED_CANDIDATES_REJECTED" {
		t.Fatal("source budget acquired full-space authority", a, explanation)
	}
	result.Report.RecordAssembly.Schema = recordAssemblySchema
	if _, err := RealizeSourceAssembly(context.Background(), "scalar.gooo", []byte(source), result); err == nil {
		t.Fatal("scalar result replayed under a record schema")
	}
	source = strings.Replace(scalarCriteriaSource, `criteria_input "[17]"`, `criteria_input "[17]" criteria_input "[19]"`, 1)
	q := generateScalarCriteria(t, source).Report.RecordAssembly.CriteriaQuestions
	if q.Status != "ROUND_BUDGET" || q.EffectiveInputs != 2 || len(q.AcquiredInputs) != 1 {
		t.Fatal("round budget became a claim about the unacquired input", q)
	}
}
