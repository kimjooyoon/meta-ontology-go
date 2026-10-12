package assemblyspec

import "fmt"

// Criteria questions have typed inputs and the containing source predicates as
// their obligations. Other observation modes keep their separate oracle path.
func (s Spec) validateCriteriaQuestions() error {
	o := s.Observation
	if s.Criteria == nil || len(o.Inputs)+len(o.ValueInputs) != 0 || o.OracleActivity != "" {
		return fmt.Errorf("observe criteria_input requires criteria and cannot mix input modes or a record oracle")
	}
	seen := make(map[string]bool, len(o.CriteriaInputs))
	held := make(map[string]bool)
	for _, input := range s.Criteria.HoldoutInputs {
		key, _ := RecordValueKey(input)
		held[key] = true
	}
	for _, c := range s.ValueHoldoutCases {
		key, _ := RecordValueKey(c.Inputs)
		held[key] = true
	}
	for _, input := range o.CriteriaInputs {
		key, err := criterionInputKey(input)
		if err != nil || seen[key] {
			return fmt.Errorf("criteria question must be a unique canonical positional tuple: %s", input)
		}
		if held[key] {
			return fmt.Errorf("criteria question input is reserved for held evaluation: %s", input)
		}
		seen[key] = true
	}
	return nil
}
