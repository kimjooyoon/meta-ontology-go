package valuecatalog

import "github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"

var catalogInputs = []int64{-2, 0, 41}

func executeProgram(program valueexecution.Program, delta int64) ProgramResult {
	result := ProgramResult{
		Activity: program.Activity, Program: program.Text, Operation: program.Operation,
		SemanticFingerprint: program.SemanticFingerprint,
	}
	for _, input := range catalogInputs {
		actual, err := program.Execute([]int64{input})
		item := CaseResult{Input: input, Expected: input + delta, Actual: actual, Passed: err == nil && actual == input+delta}
		if err != nil {
			item.Reason = valueexecution.Reason(err)
		}
		if item.Passed {
			result.Passed++
		}
		result.Cases = append(result.Cases, item)
	}
	return result
}
