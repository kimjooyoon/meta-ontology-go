package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

func (graph compositionGraph) includeCriteriaLocalInputs(seen *compositionSelectionInputs, activityID string,
	evidence *bodycodegen.RecordCriteriaEvidence) error {
	if evidence == nil {
		return nil
	}
	if err := graph.includeLocalCallObservation(seen, activityID, evidence.LocalCalls); err != nil {
		return err
	}
	for _, input := range evidence.Inputs {
		for _, check := range input.Checks {
			if check.LocalCalls == nil {
				continue
			}
			if err := graph.includeLocalCallObservation(seen, check.LocalCalls.ActivityID, check.LocalCalls); err != nil {
				return err
			}
		}
	}
	return nil
}
