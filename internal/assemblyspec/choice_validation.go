package assemblyspec

import "fmt"

func (s Spec) fixedValueContract() bool {
	return len(s.Choices) == 0 && len(s.ValueCases) > 0 && len(s.Cases)+len(s.HoldoutCases) == 0 &&
		s.Criteria == nil && s.Observation == nil && s.Search == nil && s.FillPlan == nil && s.Seed == "" &&
		len(s.ConditionCases) == 0 && len(s.Picked) == 0
}

func (s Spec) validateChoices() error {
	if len(s.Choices) == 0 && len(s.ValueCases) > 0 && !s.fixedValueContract() {
		return fmt.Errorf("fixed value_case cannot combine with Integer cases, seed, criteria, observe or picked choices")
	}
	if len(s.Choices) > 16 || len(s.Choices) == 0 && !s.fixedValueContract() {
		return fmt.Errorf("assembly requires 1..16 choices")
	}
	if err := s.validatePathHoldout(); err != nil {
		return err
	}
	var seen [16]string
	for i, c := range s.Choices {
		duplicate := false
		for _, id := range seen[:i] {
			duplicate = duplicate || id == c.ID
		}
		if !boundedText(c.ID, 128) || duplicate || !boundedText(c.Intent, 512) ||
			c.Occurrence < 0 || c.Occurrence > 127 {
			return fmt.Errorf("assembly choice %q needs a unique ID, intent and occurrence in 0..127", c.ID)
		}
		seen[i] = c.ID
		switch c.Kind {
		case "operand_order", "branch_layout", "root_order":
			if c.Alternative != "" {
				return fmt.Errorf("assembly layout choice %q cannot name an alternative local", c.ID)
			}
		case "local_reference", "assignment_target":
			if !boundedText(c.Alternative, 128) {
				return fmt.Errorf("assembly name choice %q requires an alternative local", c.ID)
			}
		case "field_value", "field_update":
			if !boundedText(c.Alternative, 512) || len(s.Cases) != 0 || (len(s.ValueCases) == 0 && s.Criteria == nil) || len(s.Choices) > 6 {
				return fmt.Errorf("field assembly requires an alternative expression, 1..6 choices and value_case or criteria obligations")
			}
		default:
			return fmt.Errorf("unknown assembly choice kind %q", c.Kind)
		}
	}
	return nil
}
