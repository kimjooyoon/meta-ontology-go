package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func loadRuntimePlanContract(reader SourceReader, filename string, source []byte, plan valueexecution.Plan) (string, error) {
	raw, err := reader.ReadFile(filename)
	if err != nil {
		return "", valueexecution.Failure{
			Code: valueexecution.ReasonSourceReadFailed, Stage: "PLAN", Step: "read-runtime-plan", Detail: err.Error(),
		}
	}
	digest, err := validateRuntimePlanContract(source, plan, raw)
	if err != nil {
		return "", valueexecution.Failure{
			Code: valueexecution.ReasonPlanInvalid, Stage: "PLAN", Step: "validate-runtime-plan-contract", Detail: err.Error(),
		}
	}
	return digest, nil
}
