package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func loadRuntimePlanContract(reader SourceReader, filename string, source []byte, plan valueexecution.Plan) (string, error) {
	_, digest, err := loadRuntimePlanExecutionEvidence(reader, filename, source, plan)
	return digest, err
}

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
	digest, err := validateRuntimePlanContract(source, plan, raw)
	if err != nil {
		return runtimePlanDocument{}, "", valueexecution.Failure{
			Code: valueexecution.ReasonPlanInvalid, Stage: "PLAN", Step: "validate-runtime-plan-contract", Detail: err.Error(),
		}
	}
	return document, digest, nil
}
