package bodycodegen

import "context"

// Prepare the fixed predicate programs before evaluating any candidate input.
func prepareCriteriaEvaluators(ctx context.Context, criteria []preparedRecordCriterion) ([]*criteriaTupleEvaluator, error) {
	evaluators := make([]*criteriaTupleEvaluator, len(criteria))
	for i, predicate := range criteria {
		program := predicate.program
		var err error
		evaluators[i], err = newCriteriaTupleEvaluator(ctx, []byte(program.Source), program.Report.Activity, program.Report.RecordTypes, nil)
		if err != nil {
			return nil, err
		}
	}
	return evaluators, nil
}
