package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func loadRuntimePlanExecutionEvidence(
	reader SourceReader,
	filename string,
	source []byte,
	plan valueexecution.Plan,
) (runtimePlanDocument, string, error) {
	raw, err := reader.ReadFile(filename)
	if err != nil {
		return runtimePlanDocument{}, "", valueexecution.Failure{
			Code: valueexecution.ReasonSourceReadFailed, Stage: "PLAN", Step: "read-runtime-plan", Detail: err.Error(),
		}
	}
	document, err := decodeRuntimePlanContract(raw)
	if err != nil {
		return runtimePlanDocument{}, "", valueexecution.Failure{
			Code: valueexecution.ReasonPlanInvalid, Stage: "PLAN", Step: "validate-runtime-plan-contract", Detail: err.Error(),
		}
	}
	if err := validateRuntimePlanContract(source, plan, document); err != nil {
		return runtimePlanDocument{}, "", valueexecution.Failure{
			Code: valueexecution.ReasonPlanInvalid, Stage: "PLAN", Step: "validate-runtime-plan-contract", Detail: err.Error(),
		}
	}
	return document, "sha256:" + cache.HashBytes(raw).String(), nil
}
