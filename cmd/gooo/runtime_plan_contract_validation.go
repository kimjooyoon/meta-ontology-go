package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func validateRuntimePlanContract(source []byte, plan valueexecution.Plan, document runtimePlanDocument) error {
	file, err := validateRuntimePlanIdentity(source, plan, document)
	if err != nil {
		return err
	}
	return validateRuntimePlanTypedPlan(file, document)
}
