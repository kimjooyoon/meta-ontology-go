package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func criteriaValueCases(s *assemblyspec.Spec, held bool) []assemblyspec.ValueCase {
	if !assemblyspec.IsScalarCriteria(s) {
		if held {
			return s.ValueHoldoutCases
		}
		return s.ValueCases
	}
	declared := s.Cases
	if held {
		declared = s.HoldoutCases
	}
	var cases []assemblyspec.ValueCase
	for _, c := range declared {
		cases = append(cases, assemblyspec.ValueCase{Inputs: fmt.Sprintf("[%d]", c.Input), Expected: fmt.Sprint(c.Expected)})
	}
	return cases
}

func evaluateScalarCriteriaCases(ctx context.Context, source []byte, activity string, records []RecordType,
	cases []assemblyspec.ValueCase, calls *localCallRecorder) ([]RecordAssemblyCase, error) {
	e, err := newCriteriaTupleEvaluator(ctx, source, activity, records, calls)
	if err != nil {
		return nil, err
	}
	var results []RecordAssemblyCase
	for i, c := range cases {
		expected, err := decodeRecordCaseValue([]byte(c.Expected), e.output, records)
		if err != nil {
			return nil, err
		}
		wanted, err := encodeCriteriaOutput(expected, e.output, records)
		if err != nil {
			return nil, err
		}
		actual, err := e.evaluate(c.Inputs, i)
		if err != nil {
			return nil, err
		}
		results = append(results, RecordAssemblyCase{Inputs: json.RawMessage(c.Inputs),
			Expected: wanted, Actual: actual, Passed: string(actual) == string(wanted)})
	}
	return results, ctx.Err()
}
