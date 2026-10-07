package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type AssemblyFeedback struct {
	Inputs   []json.RawMessage
	Expected json.RawMessage
}

// ExtendAssemblyCases promotes explicit caller feedback into a new source
// revision. Matching holdouts become selection cases; conflicting answers fail.
func ExtendAssemblyCases(ctx context.Context, filename string, source []byte, activity string, feedback []AssemblyFeedback) ([]byte, int, error) {
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return nil, 0, err
	}
	if spec == nil || spec.FillPlan != nil || len(feedback) > 128 {
		return nil, 0, fmt.Errorf("feedback requires a choice/search contract and at most 128 cases")
	}
	added := 0
	for _, row := range feedback {
		var changed bool
		if len(spec.ValueCases)+len(spec.ValueHoldoutCases) > 0 {
			changed, err = extendValueCase(spec, row)
		} else {
			changed, err = extendIntegerCase(spec, row)
		}
		if err != nil {
			return nil, 0, err
		}
		if changed {
			added++
		}
	}
	if err := spec.Validate(); err != nil {
		return nil, 0, err
	}
	if added == 0 {
		return append([]byte(nil), source...), 0, nil
	}
	next, err := rewriteAssemblyContract(filename, source, activity, spec)
	return next, added, err
}

func extendIntegerCase(spec *assemblyspec.Spec, row AssemblyFeedback) (bool, error) {
	if len(row.Inputs) != 1 {
		return false, fmt.Errorf("integer feedback requires one input")
	}
	input, err := strconv.ParseInt(string(row.Inputs[0]), 10, 64)
	if err != nil {
		return false, err
	}
	expected, err := strconv.ParseInt(string(row.Expected), 10, 64)
	if err != nil {
		return false, err
	}
	for _, test := range append(append([]assemblyspec.Case(nil), spec.Cases...), spec.HoldoutCases...) {
		if test.Input == input && test.Expected != expected {
			return false, fmt.Errorf("feedback conflicts with an existing integer expectation")
		}
	}
	for _, test := range spec.Cases {
		if test.Input == input {
			return false, nil
		}
	}
	holdouts := spec.HoldoutCases[:0]
	for _, test := range spec.HoldoutCases {
		if test.Input != input {
			holdouts = append(holdouts, test)
		}
	}
	spec.HoldoutCases = holdouts
	spec.Cases = append(spec.Cases, assemblyspec.Case{Input: input, Expected: expected})
	return true, nil
}

func extendValueCase(spec *assemblyspec.Spec, row AssemblyFeedback) (bool, error) {
	if len(row.Inputs) == 0 || len(row.Inputs) > 16 {
		return false, fmt.Errorf("value feedback requires 1..16 inputs")
	}
	raw, err := json.Marshal(row.Inputs)
	if err != nil {
		return false, err
	}
	input, err := canonicalFeedbackJSON(raw)
	if err != nil {
		return false, err
	}
	expected, err := canonicalFeedbackJSON(row.Expected)
	if err != nil {
		return false, err
	}
	for _, test := range append(append([]assemblyspec.ValueCase(nil), spec.ValueCases...), spec.ValueHoldoutCases...) {
		key, _ := canonicalFeedbackJSON([]byte(test.Inputs))
		answer, _ := canonicalFeedbackJSON([]byte(test.Expected))
		if key == input && answer != expected {
			return false, fmt.Errorf("feedback conflicts with an existing value expectation")
		}
	}
	for _, test := range spec.ValueCases {
		key, _ := canonicalFeedbackJSON([]byte(test.Inputs))
		if key == input {
			return false, nil
		}
	}
	holdouts := spec.ValueHoldoutCases[:0]
	for _, test := range spec.ValueHoldoutCases {
		key, _ := canonicalFeedbackJSON([]byte(test.Inputs))
		if key != input {
			holdouts = append(holdouts, test)
		}
	}
	spec.ValueHoldoutCases = holdouts
	spec.ValueCases = append(spec.ValueCases, assemblyspec.ValueCase{Inputs: input, Expected: expected})
	return true, nil
}

func canonicalFeedbackJSON(raw []byte) (string, error) {
	if !json.Valid(raw) {
		return "", fmt.Errorf("feedback must contain one JSON value")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(value)
	return string(canonical), err
}
