package bodycodegen

import (
	"context"
	"fmt"
)

func scoreSourceFillCandidate(ctx context.Context, filename string, source []byte, generated, activity string,
	plan IRBodyFillPlan, r *FillCandidate) error {
	if plan.Schema != bodyFillRecordPlanSchema {
		var err error
		r.CaseResults, r.TestCasesPassed, err = evaluateIntegerCasesContext(ctx, []byte(generated), activity, plan.TestCases)
		if err != nil {
			return err
		}
		r.HoldoutCaseResults, r.HoldoutCasesPassed, err = evaluateIntegerCasesContext(ctx, []byte(generated), activity, plan.HoldoutTestCases)
		r.TestCasesTotal, r.HoldoutCasesTotal = len(plan.TestCases), len(plan.HoldoutTestCases)
		return err
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if file == nil || diagnostics.HasErrors() {
		return fmt.Errorf("parse completed record fill: %v", diagnostics)
	}
	_, records, err := resolveBodyModel(file)
	if err != nil {
		return err
	}
	r.ValueCaseResults, err = evaluateRecordAssembly(ctx, []byte(generated), activity, records, plan.ValueCases)
	if err != nil {
		return err
	}
	r.ValueHoldoutResults, err = evaluateRecordAssembly(ctx, []byte(generated), activity, records, plan.ValueHoldoutCases)
	if err != nil {
		return err
	}
	r.TestCasesTotal, r.HoldoutCasesTotal = len(r.ValueCaseResults), len(r.ValueHoldoutResults)
	for _, row := range r.ValueCaseResults {
		if row.Passed {
			r.TestCasesPassed++
		}
	}
	for _, row := range r.ValueHoldoutResults {
		if row.Passed {
			r.HoldoutCasesPassed++
		}
	}
	return nil
}
