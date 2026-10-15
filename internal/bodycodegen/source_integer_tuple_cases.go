package bodycodegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Source value_case tuples can describe Integer arguments directly. Keep the
// existing integer evaluator and decision receipt, including exact int64 values.
func sourceIntegerTupleCases(filename string, source []byte, name string, spec *assemblyspec.Spec) ([]IRBodyFillTestCase, []IRBodyFillTestCase, bool, error) {
	if len(spec.ValueCases) == 0 {
		return nil, nil, false, nil
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() {
		return nil, nil, false, diagnostics.Error()
	}
	var selected *syntax.ActivityDecl
	for _, decl := range file.Declarations {
		if a, ok := decl.(*syntax.ActivityDecl); ok && a.Name == name {
			selected = a
		}
	}
	if selected == nil {
		return nil, nil, false, fmt.Errorf("source fill activity %q is absent", name)
	}
	if sourceScalarKind(file, selected.Output) != "Integer" {
		return nil, nil, false, nil
	}
	for _, input := range selected.Inputs {
		if sourceScalarKind(file, input.Name) != "Integer" {
			return nil, nil, false, nil
		}
	}
	arity := len(selected.Inputs)
	if arity < 1 || arity > 16 {
		return nil, nil, true, fmt.Errorf("Integer tuple fill requires 1..16 inputs")
	}
	decode := func(cases []assemblyspec.ValueCase) ([]IRBodyFillTestCase, error) {
		var out []IRBodyFillTestCase
		for i, row := range cases {
			var values []json.RawMessage
			if err := json.Unmarshal([]byte(row.Inputs), &values); err != nil || len(values) != arity {
				return nil, fmt.Errorf("Integer tuple case %d requires %d positional inputs", i, arity)
			}
			inputs := make([]int64, len(values))
			var expected int64
			for p, value := range values {
				if string(value) == "null" {
					return nil, fmt.Errorf("Integer tuple case %d input %d is null", i, p)
				}
				if err := json.Unmarshal(value, &inputs[p]); err != nil {
					return nil, fmt.Errorf("Integer tuple case %d input %d requires int64: %w", i, p, err)
				}
			}
			if strings.TrimSpace(row.Expected) == "null" {
				return nil, fmt.Errorf("Integer tuple case %d expected value is null", i)
			}
			if err := json.Unmarshal([]byte(row.Expected), &expected); err != nil {
				return nil, fmt.Errorf("Integer tuple case %d expected value requires int64: %w", i, err)
			}
			out = append(out, IRBodyFillTestCase{Input: inputs[0], Inputs: inputs, Expected: expected})
		}
		return out, nil
	}
	cases, err := decode(spec.ValueCases)
	if err != nil {
		return nil, nil, true, err
	}
	holdout, err := decode(spec.ValueHoldoutCases)
	return cases, holdout, true, err
}
