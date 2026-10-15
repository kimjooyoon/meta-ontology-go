package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

func (graph compositionGraph) includeCriteriaQuestionCalls(seen *compositionSelectionInputs, activityID string, r *bodycodegen.CriteriaQuestionReceipt) error {
	if r == nil {
		return nil
	}
	for _, output := range r.Outputs {
		if err := graph.includeLocalCallObservation(seen, activityID, output.LocalCalls); err != nil {
			return err
		}
	}
	for _, evaluation := range r.Evaluations {
		for _, check := range evaluation.Result.Checks {
			if check.LocalCalls != nil {
				if err := graph.includeLocalCallObservation(seen, check.LocalCalls.ActivityID, check.LocalCalls); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
