package generation

import "fmt"

func validatePlanRefutedEvidence(plan Plan) error {
	if !validOrderedIndicatorIDs(plan.RefutedIndicatorIDs) {
		return fmt.Errorf("refuted indicator IDs are not canonical")
	}
	seen := make(map[string]struct{}, len(plan.Counterexamples))
	refuted := make(map[string]struct{}, len(plan.RefutedIndicatorIDs))
	for _, indicatorID := range plan.RefutedIndicatorIDs {
		refuted[indicatorID] = struct{}{}
	}
	for _, counterexample := range plan.Counterexamples {
		if counterexample.ID == "" || counterexample.IndicatorID == "" ||
			counterexample.BlockerID == "" || counterexample.Stage == "" ||
			counterexample.Step == "" || counterexample.Reason == "" ||
			counterexample.UnknownClass == "" || counterexample.NextOperation == "" {
			return fmt.Errorf("incomplete planner counterexample")
		}
		if !validActionIndicatorID(counterexample.IndicatorID) {
			return fmt.Errorf("counterexample indicator ID is malformed")
		}
		if _, exists := refuted[counterexample.IndicatorID]; !exists {
			return fmt.Errorf("counterexample is not linked to a refuted indicator")
		}
		if _, duplicate := seen[counterexample.ID]; duplicate {
			return fmt.Errorf("duplicate planner counterexample %q", counterexample.ID)
		}
		seen[counterexample.ID] = struct{}{}
	}
	return nil
}
