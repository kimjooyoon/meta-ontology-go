package assemblyspec

import "fmt"

// RecordValueKey compares the exact integer, text, Boolean and presence values
// in typed record cases. Source JSON spelling remains available separately.
func RecordValueKey(raw string) (string, error) {
	if _, err := CanonicalValue(raw); err != nil {
		return "", err
	}
	return recordHoldoutInputKey(raw)
}

func (s Spec) validateRecordObservationInputs() error {
	o := s.Observation
	if len(s.Cases) != 0 || len(o.Inputs) != 0 || len(o.ValueInputs) == 0 {
		return fmt.Errorf("record observe requires value_input positional tuples and value_case declarations")
	}
	held := make(map[string]bool, len(s.ValueHoldoutCases))
	for _, c := range s.ValueHoldoutCases {
		key, err := recordHoldoutInputKey(c.Inputs)
		if err != nil {
			return err
		}
		held[key] = true
	}
	seen := make(map[string]bool, len(o.ValueInputs))
	for _, input := range o.ValueInputs {
		canonical, err := CanonicalValue(input)
		if err != nil || canonical != input {
			return fmt.Errorf("observe value_input requires bounded canonical JSON")
		}
		key, err := recordHoldoutInputKey(input)
		if err != nil {
			return err
		}
		if seen[key] {
			return fmt.Errorf("duplicate observe value_input %s", input)
		}
		if held[key] {
			return fmt.Errorf("observe value_input %s is reserved by holdout_value_case", input)
		}
		seen[key] = true
	}
	return nil
}
