package assemblyspec

import (
	"encoding/json"
	"fmt"
)

// Criteria binds named pure predicates to the containing assembly. Predicate
// signatures are the assembly inputs followed by its result, returning Boolean.
// Selection and held input tuples have no manufactured expected record value.
type Criteria struct {
	Predicates    []Criterion `json:"predicates"`
	Inputs        []string    `json:"inputs"`
	HoldoutInputs []string    `json:"holdout_inputs,omitempty"`
}

type Criterion struct {
	ID       string `json:"id"`
	Activity string `json:"activity"`
}

func (c *Criteria) Clone() *Criteria {
	if c == nil {
		return nil
	}
	copy := *c
	copy.Predicates = append([]Criterion(nil), c.Predicates...)
	copy.Inputs = append([]string(nil), c.Inputs...)
	copy.HoldoutInputs = append([]string(nil), c.HoldoutInputs...)
	return &copy
}

func (s Spec) criteriaInputCount() int {
	if s.Criteria == nil {
		return 0
	}
	return len(s.Criteria.Inputs) + len(s.Criteria.HoldoutInputs)
}

func (s Spec) validateCriteria() error {
	c := s.Criteria
	if c == nil {
		return nil
	}
	if len(c.Predicates) < 1 || len(c.Predicates) > 8 || len(c.Inputs) < 1 ||
		s.criteriaInputCount() > 128 || s.Search != nil || s.FillPlan != nil ||
		len(s.Cases)+len(s.ConditionCases)+len(s.HoldoutCases) != 0 {
		return fmt.Errorf("criteria requires 1..8 predicates, 1..128 typed input tuples and record field choices")
	}
	for _, choice := range s.Choices {
		if choice.Kind != "field_value" && choice.Kind != "field_update" {
			return fmt.Errorf("criteria requires record field choices")
		}
	}
	seen := make(map[string]bool, len(c.Predicates))
	for _, predicate := range c.Predicates {
		if !boundedText(predicate.ID, 128) || !boundedText(predicate.Activity, 256) || seen[predicate.ID] {
			return fmt.Errorf("criteria predicate requires a unique ID and activity")
		}
		seen[predicate.ID] = true
	}
	return s.validateCriteriaInputs()
}

func (s Spec) validateCriteriaInputs() error {
	selection := make(map[string]bool)
	for _, c := range s.ValueCases {
		key, err := RecordValueKey(c.Inputs)
		if err != nil {
			return err
		}
		selection[key] = true
	}
	if s.Observation != nil {
		for _, input := range s.Observation.ValueInputs {
			key, err := RecordValueKey(input)
			if err != nil {
				return err
			}
			selection[key] = true
		}
	}
	seen := make(map[string]bool)
	for _, input := range s.Criteria.Inputs {
		key, err := criterionInputKey(input)
		if err != nil || seen[key] {
			return fmt.Errorf("criteria input must be a unique canonical positional tuple: %s", input)
		}
		selection[key], seen[key] = true, true
	}
	var held []string
	for _, c := range s.ValueHoldoutCases {
		held = append(held, c.Inputs)
	}
	held = append(held, s.Criteria.HoldoutInputs...)
	for _, input := range held {
		key, err := criterionInputKey(input)
		if err != nil || selection[key] {
			return fmt.Errorf("criteria holdout input must be canonical and separate from selection and other holdouts: %s", input)
		}
		selection[key] = true
	}
	return nil
}

func criterionInputKey(input string) (string, error) {
	canonical, err := CanonicalValue(input)
	if err != nil || canonical != input {
		return "", fmt.Errorf("criteria input requires canonical JSON")
	}
	var values []json.RawMessage
	if err := json.Unmarshal([]byte(input), &values); err != nil || len(values) < 1 || len(values) > 16 {
		return "", fmt.Errorf("criteria input requires 1..16 positional values")
	}
	return RecordValueKey(input)
}
