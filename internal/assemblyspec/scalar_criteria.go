package assemblyspec

import "fmt"

// Scalar criteria retain the existing path choices and original scalar signature.
// No expected output is invented for an input that declares only predicates.
func IsScalarCriteria(s *Spec) bool {
	if s == nil || s.Criteria == nil || len(s.Choices) == 0 {
		return false
	}
	return s.Choices[0].Kind != "field_value" && s.Choices[0].Kind != "field_update"
}

func (s Spec) validateCriteriaChoiceKinds() error {
	scalar := IsScalarCriteria(&s)
	for _, choice := range s.Choices {
		if scalar && (choice.Kind == "local_reference" || choice.Kind == "assignment_target") && !identifier(choice.Alternative) {
			return fmt.Errorf("scalar criteria local alternatives must be identifiers")
		}
		field := choice.Kind == "field_value" || choice.Kind == "field_update"
		if scalar == field {
			return fmt.Errorf("criteria cannot mix scalar paths and record field choices")
		}
	}
	if scalar && len(s.ValueCases)+len(s.ValueHoldoutCases) != 0 {
		return fmt.Errorf("scalar criteria use case and holdout_case for exact Integer outputs")
	}
	if !scalar && len(s.Cases)+len(s.HoldoutCases) != 0 {
		return fmt.Errorf("record criteria use value_case and holdout_value_case")
	}
	return nil
}
