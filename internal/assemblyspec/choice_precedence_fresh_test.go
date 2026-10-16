package assemblyspec

import "testing"

func TestChoiceValidationOrderFresh1761(t *testing.T) {
	for _, mode := range []string{"missing-count-before-holdout", "holdout-before-malformed-choice"} {
		t.Run(mode, func(t *testing.T) {
			spec := Spec{MaxAttempts: 1, Cases: []Case{{Input: 1, Expected: 1}},
				HoldoutCases: []Case{{Input: 1, Expected: 1}}}
			expected := "assembly requires 1..16 choices"
			if mode == "holdout-before-malformed-choice" {
				spec.Choices = []Choice{{ID: "", Kind: "operand_order"}}
				expected = "typed path holdout input 1 overlaps a selection case or another holdout"
			}
			err := spec.Validate()
			if err == nil || err.Error() != expected {
				t.Fatalf("known validation priority changed: %v, want %s", err, expected)
			}
		})
	}
}
