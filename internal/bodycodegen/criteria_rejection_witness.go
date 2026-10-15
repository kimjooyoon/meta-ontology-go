package bodycodegen

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func criteriaRejectionWitness(a *RecordAssemblyReceipt, c CriteriaQuestionCandidate) (CriteriaRejectionGroup, bool) {
	q := a.CriteriaQuestions
	for _, index := range c.ExactCaseOutputs {
		output, ok := criteriaExplanationOutput(q, index, c.Mask)
		if !ok {
			continue
		}
		if expected, ok := criteriaExactRejection(a.Cases, output); ok {
			return CriteriaRejectionGroup{Kind: "EXACT_CASE", InputJSON: output.Input, ExpectedJSON: expected,
				Witnesses: []CriteriaRejectionWitness{{Mask: c.Mask, ActualJSON: output.Value, OutputIndex: index}}}, true
		}
	}
	for _, index := range c.CriteriaIndices {
		if group, ok := criteriaPredicateRejection(q, c.Mask, index); ok {
			return group, true
		}
	}
	return CriteriaRejectionGroup{}, false
}

func criteriaPredicateRejection(q *CriteriaQuestionReceipt, mask uint16, index int) (CriteriaRejectionGroup, bool) {
	if index < 0 || index >= len(q.Evaluations) {
		return CriteriaRejectionGroup{}, false
	}
	e := q.Evaluations[index]
	o, ok := criteriaExplanationOutput(q, e.OutputIndex, mask)
	if !ok || e.Mask != mask || e.Result.Error != "" || e.Result.InputJSON != o.Input || e.Result.ActualJSON != o.Value ||
		(!slices.Contains(q.InitialInputs, o.Input) && !slices.Contains(q.AcquiredInputs, o.Input)) {
		return CriteriaRejectionGroup{}, false
	}
	for _, check := range e.Result.Checks {
		if check.Status != "UNSATISFIED" {
			continue
		}
		if binding := criteriaExplanationBinding(q, check.ID); binding != nil {
			return CriteriaRejectionGroup{Kind: "CRITERION", Criterion: binding, InputJSON: o.Input,
				Witnesses: []CriteriaRejectionWitness{{Mask: mask, ActualJSON: o.Value, OutputIndex: e.OutputIndex, EvaluationIndex: new(index)}}}, true
		}
	}
	return CriteriaRejectionGroup{}, false
}

func criteriaExactRejection(cases []RecordAssemblyCase, output CriteriaQuestionOutput) (string, bool) {
	input, err := assemblyspec.RecordValueKey(output.Input)
	if err != nil {
		return "", false
	}
	actual, err := assemblyspec.RecordValueKey(output.Value)
	if err != nil {
		return "", false
	}
	for _, c := range cases {
		key, err := assemblyspec.RecordValueKey(string(c.Inputs))
		if err != nil || key != input {
			continue
		}
		expected, err := assemblyspec.RecordValueKey(string(c.Expected))
		return expected, err == nil && expected != actual
	}
	return "", false
}
