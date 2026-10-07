package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

func replaySourceIRSearchPlan(ctx context.Context, filename string, source []byte, prior Result) (IRBodySearchPlan, error) {
	spec, err := SourceAssembly(ctx, filename, source, prior.Report.Activity)
	if err != nil {
		return IRBodySearchPlan{}, err
	}
	plan, err := sourceIRSearchPlan(spec)
	if err != nil {
		return plan, fmt.Errorf("IR search replay requires its source-declared contract: %w", err)
	}
	r := prior.Report
	if r.Schema != schema || r.Decision != "PASS" || !r.TypecheckPassed || !r.DeterministicReplay ||
		r.BodyPaths != nil || r.BodyFill != nil || r.RecordAssembly != nil || prior.GoooSource != "" ||
		r.PlanSHA256 != completenessPlanSHA(r) {
		return plan, fmt.Errorf("IR search projection profile is missing or changed")
	}
	generated, err := generateIRBodySearchCandidates(&plan)
	if err != nil {
		return plan, err
	}
	encoded, err := json.Marshal(plan)
	search := r.BodySearch
	if err != nil || search.Schema != bodySearchPlanSchema || search.IRPlanSHA256 != digest(encoded) ||
		search.CandidateCount != len(plan.Candidates) || !reflect.DeepEqual(search.CandidateGeneration, generated) {
		return plan, fmt.Errorf("IR search plan or candidate space differs from the Gooo source")
	}
	candidate, found := candidateByID(plan.Candidates, search.SelectedCandidateID)
	if !found || candidate.Expression != search.SelectedExpression {
		return plan, fmt.Errorf("IR search selection is outside the source-declared candidate set")
	}
	return plan, nil
}

func replayIRSearchFiniteEvidence(ctx context.Context, prior Result, plan IRBodySearchPlan) error {
	s := prior.Report.BodySearch
	suites := []struct {
		name, digest, failure string
		cases                 []IRBodyFillTestCase
		results               []IRBodyFillCaseResult
		passed, total         int
		accuracy              *float64
	}{
		{name: "training", digest: s.TrainingSuiteSHA256, cases: plan.TestCases, results: s.TrainingCaseResults,
			passed: s.TrainingPassed, total: s.TrainingTotal, accuracy: s.TrainingAccuracyPercent},
		{name: "holdout", digest: s.HoldoutSuiteSHA256, failure: s.HoldoutError,
			cases: plan.HoldoutTestCases, results: s.HoldoutCaseResults,
			passed: s.HoldoutPassed, total: s.HoldoutTotal, accuracy: s.HoldoutAccuracyPercent},
	}
	for _, suite := range suites {
		var results []IRBodyFillCaseResult
		var accuracy *float64
		passed, suiteDigest, failure := 0, "", ""
		if len(suite.cases) > 0 {
			encoded, _ := json.Marshal(suite.cases)
			suiteDigest = digest(encoded)
			var err error
			results, passed, err = evaluateIntegerCasesContext(ctx, []byte(prior.Source), prior.Report.Activity, suite.cases)
			if err != nil {
				failure = err.Error()
			} else {
				value := float64(passed) * 100 / float64(len(suite.cases))
				accuracy = &value
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if suite.total != len(suite.cases) || suite.digest != suiteDigest || suite.passed != passed ||
			suite.failure != failure || !reflect.DeepEqual(suite.accuracy, accuracy) || !equalIRBodyFillCaseResults(suite.results, results) {
			return fmt.Errorf("IR search %s observations differ from the source-declared cases", suite.name)
		}
	}
	return nil
}
