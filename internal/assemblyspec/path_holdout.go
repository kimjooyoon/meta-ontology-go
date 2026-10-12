package assemblyspec

import "fmt"

func (s Spec) validatePathHoldout() error {
	if len(s.HoldoutCases) == 0 {
		return nil
	}
	if len(s.Cases) == 0 || len(s.ValueCases)+len(s.ValueHoldoutCases) != 0 {
		return fmt.Errorf("typed path holdout_case requires initial Integer case declarations")
	}
	used := make(map[int64]bool, len(s.Cases)+len(s.ConditionCases))
	for _, c := range s.Cases {
		used[c.Input] = true
	}
	for _, c := range s.ConditionCases {
		used[c.Input] = true
	}
	held := make(map[int64]bool, len(s.HoldoutCases))
	for _, c := range s.HoldoutCases {
		if used[c.Input] || held[c.Input] {
			return fmt.Errorf("typed path holdout input %d overlaps a selection case or another holdout", c.Input)
		}
		held[c.Input] = true
	}
	return nil
}
