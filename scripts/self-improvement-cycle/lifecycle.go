package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/cyclecontext"

func classifyCycle(conclusion string, bindingsValid bool) cyclecontext.R_State {
	return cyclecontext.Evaluate(cyclecontext.R_Observation{
		F_conclusion: conclusion, F_bindings_valid: bindingsValid,
	})
}
