package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func validateRuntimePlanContract(source []byte, plan valueexecution.Plan, raw []byte) (string, error) {
	document, err := decodeRuntimePlanContract(raw)
	if err != nil {
		return "", err
	}
	file, err := validateRuntimePlanIdentity(source, plan, document)
	if err != nil {
		return "", err
	}
	if err := validateRuntimePlanTypedPlan(file, document); err != nil {
		return "", err
	}
	return "sha256:" + cache.HashBytes(raw).String(), nil
}
