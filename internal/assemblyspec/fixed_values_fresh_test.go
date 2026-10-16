package assemblyspec

import "testing"

func TestFixedValueBoundaryFresh1755(t *testing.T) {
	for _, mode := range []string{"fixed", "checkpoint", "budget64", "no-cases", "holdout-only", "integer-case",
		"seed", "criteria", "observe", "picked", "zero-budget", "large-budget", "too-many-cases", "too-many-choices"} {
		t.Run(mode, func(t *testing.T) {
			spec := Spec{MaxAttempts: 1, ValueCases: []ValueCase{{Inputs: `[false]`, Expected: `{"enabled":false}`}}}
			mutateFixedValue1755(&spec, mode)
			err := spec.Validate()
			valid := mode == "fixed" || mode == "checkpoint" || mode == "budget64"
			if (err == nil) != valid {
				t.Fatalf("fixed contract accepted=%v want=%v: %v", err == nil, valid, err)
			}
		})
	}
}

func mutateFixedValue1755(spec *Spec, mode string) {
	switch mode {
	case "checkpoint":
		spec.Baseline = "return Gate{enabled: false}"
	case "budget64":
		spec.MaxAttempts = 64
	case "no-cases":
		spec.ValueCases = nil
	case "holdout-only":
		spec.ValueHoldoutCases, spec.ValueCases = spec.ValueCases, nil
	case "integer-case":
		spec.Cases = []Case{{Input: 0, Expected: 0}}
	case "seed":
		spec.Seed = "unnecessary-draw"
	case "criteria":
		spec.Criteria = &Criteria{Predicates: []Criterion{{ID: "p", Activity: "Predicate"}}, Inputs: []string{`[false]`}}
	case "observe":
		spec.Observation = &Observation{ValueInputs: []string{`[true]`}, MaxCandidates: 1, MaxRounds: 1}
	case "picked":
		spec.Picked = []Pick{{ID: "undeclared", Label: "value_first"}}
	case "zero-budget":
		spec.MaxAttempts = 0
	case "large-budget":
		spec.MaxAttempts = 65
	case "too-many-cases":
		for len(spec.ValueCases) < 129 {
			spec.ValueCases = append(spec.ValueCases, spec.ValueCases[0])
		}
	case "too-many-choices":
		spec.Choices = make([]Choice, 17)
	}
}
