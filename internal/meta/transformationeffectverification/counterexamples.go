package transformationeffectverification

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func VerifyCounterexamples(opts Options) (Report, error) {
	bundle, err := loadBundle(opts)
	if err != nil {
		return failureReport(err), err
	}
	if err := validatePlan(bundle.Plan); err != nil {
		return failureReport(err), err
	}
	cases := []string{"missing", "duplicate", "stale"}
	results := make([]Counterexample, 0, len(cases))
	for _, kind := range cases {
		mutated := mutatePlan(bundle.Plan, kind)
		failure := validatePlan(mutated)
		typed, ok := failure.(*validationFailure)
		if !ok || typed.Decision != "REFUTED" || typed.Resolution != "EXACT" {
			return failureReport(&validationFailure{Decision: "REFUTED", Resolution: "EXACT",
				Stage: "counterexamples", Step: "verify-binding-mutation", Reason: "COUNTEREXAMPLE_NOT_REFUTED",
				Next: "report-counterexample", Blocked: []string{}, FieldPath: kind}), failureForCounterexample(kind)
		}
		results = append(results, Counterexample{ID: "executor-binding-" + kind, Decision: typed.Decision,
			Resolution: typed.Resolution, Stage: typed.Stage, Step: typed.Step, Reason: typed.Reason,
			FieldPath: typed.FieldPath, Expected: typed.Expected, Observed: typed.Observed})
	}
	return Report{Schema: verifierSchema, Decision: "PASS", Resolution: "EXACT", Stage: "counterexamples",
		Step: "verify-binding-mutations", Reason: "BINDING_COUNTEREXAMPLES_REFUTED", NextOperation: "none",
		BlockedBy: []string{}, SelectedPlanOperations: len(bundle.Plan.Selected), BoundExecutorOperations: len(bundle.Plan.Selected),
		UnboundExecutorOperations: 0, Counterexamples: results}, nil
}

func mutatePlan(plan generation.Plan, kind string) generation.Plan {
	mutated := plan
	mutated.Registry = append([]generation.Binding{}, plan.Registry...)
	if len(mutated.Selected) == 0 {
		return mutated
	}
	operation := mutated.Selected[0].Operation
	index := bindingIndex(mutated.Registry, operation)
	if index < 0 {
		return mutated
	}
	switch kind {
	case "missing":
		mutated.Registry = append(mutated.Registry[:index], mutated.Registry[index+1:]...)
	case "duplicate":
		mutated.Registry = append(mutated.Registry, mutated.Registry[index])
	case "stale":
		mutated.Registry[index].Executor = "stale-executor"
	}
	return mutated
}

func bindingIndex(registry []generation.Binding, operation sourcepolicy.Operation) int {
	for index, binding := range registry {
		if binding.Operation == operation {
			return index
		}
	}
	return -1
}

func failureForCounterexample(kind string) error {
	return fmt.Errorf("binding counterexample %s was accepted", kind)
}
