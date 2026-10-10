package assemblyspec

import "fmt"

func (s Spec) validateConditionCases() error {
	if len(s.ConditionCases) == 0 {
		return nil
	}
	if len(s.ConditionCases) > 128 || s.Search != nil || s.FillPlan != nil ||
		len(s.ValueCases)+len(s.ValueHoldoutCases) != 0 {
		return fmt.Errorf("condition_case requires typed paths and at most 128 condition cases")
	}
	kinds := make(map[string]string, len(s.Choices))
	for _, choice := range s.Choices {
		kinds[choice.ID] = choice.Kind
	}
	type key struct {
		choice string
		input  int64
	}
	seen := make(map[key]bool, len(s.ConditionCases))
	for _, test := range s.ConditionCases {
		kind := kinds[test.ChoiceID]
		if kind != "branch_layout" && kind != "operand_order" {
			return fmt.Errorf("condition_case %q requires a declared branch_layout or operand_order choice", test.ChoiceID)
		}
		k := key{test.ChoiceID, test.Input}
		if seen[k] {
			return fmt.Errorf("duplicate condition_case input for %q", test.ChoiceID)
		}
		seen[k] = true
	}
	return nil
}
