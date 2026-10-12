package assemblyspec

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (s Spec) validateRecordHoldout() error {
	if len(s.ValueHoldoutCases) == 0 {
		return nil
	}
	if len(s.ValueCases) == 0 {
		return fmt.Errorf("record holdout_value_case requires value_case selection examples")
	}
	inputs := make(map[string]bool, len(s.ValueCases)+len(s.ValueHoldoutCases))
	for _, c := range s.ValueCases {
		key, err := recordHoldoutInputKey(c.Inputs)
		if err != nil {
			return err
		}
		inputs[key] = true
	}
	for _, c := range s.ValueHoldoutCases {
		input, err := CanonicalValue(c.Inputs)
		expected, next := CanonicalValue(c.Expected)
		if err != nil || next != nil || input != c.Inputs || expected != c.Expected {
			return fmt.Errorf("record holdout_value_case requires bounded canonical JSON")
		}
		key, err := recordHoldoutInputKey(c.Inputs)
		if err != nil {
			return err
		}
		if inputs[key] {
			return fmt.Errorf("record holdout input %s overlaps a selection case or another holdout", c.Inputs)
		}
		inputs[key] = true
	}
	return nil
}

// These record inputs use int64 values. In that domain 0 and -0 are the same
// input even though source JSON preserves the original spelling. Text stays text.
func recordHoldoutInputKey(raw string) (string, error) {
	d := json.NewDecoder(bytes.NewBufferString(raw))
	d.UseNumber()
	var value any
	if err := d.Decode(&value); err != nil {
		return "", err
	}
	value, err := normalizeRecordHoldoutNumbers(value)
	if err != nil {
		return "", fmt.Errorf("record holdout inputs require exact int64 values: %w", err)
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func normalizeRecordHoldoutNumbers(value any) (any, error) {
	switch v := value.(type) {
	case json.Number:
		n, err := v.Int64()
		return n, err
	case []any:
		for i := range v {
			item, err := normalizeRecordHoldoutNumbers(v[i])
			if err != nil {
				return nil, err
			}
			v[i] = item
		}
	case map[string]any:
		for key := range v {
			item, err := normalizeRecordHoldoutNumbers(v[key])
			if err != nil {
				return nil, err
			}
			v[key] = item
		}
	}
	return value, nil
}
