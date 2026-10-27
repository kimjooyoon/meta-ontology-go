package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type CompositionCase struct {
	Inputs   map[string]json.RawMessage `json:"inputs"`
	Expected map[string]json.RawMessage `json:"expected"`
}

type CompositionCases struct {
	Schema  string                  `json:"schema"`
	Cases   []CompositionCase       `json:"cases"`
	Display *CompositionCaseDisplay `json:"-"`
}

func DecodeCompositionCases(raw []byte) (CompositionCases, error) {
	var suite CompositionCases
	if err := decodeCompositionInputData(raw, &suite); err != nil {
		return suite, err
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) < 1 || len(suite.Cases) > compositionRowsMax {
		return suite, fmt.Errorf("composition cases require schema and 1..128 cases")
	}
	return suite, nil
}

// ValidateCompositionCases checks all root inputs and named expectations without
// inference, program execution or model loading.
func ValidateCompositionCases(ctx context.Context, filename string, source []byte, suite CompositionCases) error {
	graph, err := prepareCompositionGraph(ctx, filename, source)
	if err != nil {
		return err
	}
	_, err = graph.inputRows(suite)
	return err
}

func (graph compositionGraph) inputRows(suite CompositionCases) ([][]json.RawMessage, error) {
	if err := validateCompositionInputMode(suite); err != nil {
		return nil, err
	}
	if len(suite.Cases) < 1 || len(suite.Cases) > compositionRowsMax {
		return nil, fmt.Errorf("composition cases require schema and 1..128 cases")
	}
	raw, err := json.Marshal(suite)
	if err != nil {
		return nil, err
	}
	if err := checkCompositionInputSize(raw, "normalized suite"); err != nil {
		return nil, err
	}
	if suite.Display != nil {
		declared, available := graph.displayInputNames(suite.Display)
		for c, test := range suite.Cases {
			if err := suite.Display.checkSubmittedInputs(test.Inputs, declared, available); err != nil {
				return nil, inputCaseDiagnostic(c, err)
			}
		}
	}
	rows := make([][]json.RawMessage, len(suite.Cases))
	for c, test := range suite.Cases {
		row, err := graph.inputRow(test, suite.Schema != CompositionInputsSchema, suite.Display)
		if err != nil {
			return nil, inputCaseDiagnostic(c, err)
		}
		rows[c] = row
	}
	return rows, nil
}

func (graph compositionGraph) inputRow(test CompositionCase, requireExpectation bool,
	display *CompositionCaseDisplay) ([]json.RawMessage, error) {
	if test.Inputs == nil {
		return nil, fmt.Errorf("composition inputs require an explicit object, including for zero caller inputs")
	}
	var storage [compositionLimit * compositionLimit]json.RawMessage
	roots, expected := 0, 0
	for _, node := range graph.nodes[:graph.count] {
		for _, input := range node.inputSlots() {
			key := node.inputKey(input.Port)
			shown := display.inputKey(node, input.Port)
			value, hasInput := test.Inputs[key]
			if input.From < 0 {
				if !hasInput {
					return nil, &InputDiagnostic{Stage: "INPUT_VALIDATION", Code: "MISSING_ROOT_INPUT",
						DeclaredKey: shown, Message: fmt.Sprintf("root input %q requires an explicit value", shown)}
				}
				canonical, err := graph.canonicalValue(value, input.Type)
				if err != nil {
					return nil, &InputDiagnostic{Stage: "INPUT_VALIDATION", Code: "INVALID_INPUT_VALUE",
						Key: shown, Message: fmt.Sprintf("input %q: %v", shown, err)}
				}
				storage[roots], roots = canonical, roots+1
			} else if hasInput {
				return nil, &InputDiagnostic{Stage: "INPUT_VALIDATION", Code: "BOUND_INPUT_PROVIDED",
					Key: shown, Message: fmt.Sprintf("bound input %q must receive its producer's result", shown)}
			}
		}
		if value, present := test.Expected[node.Name]; present {
			if _, err := graph.canonicalValue(value, node.OutputType); err != nil {
				shown := display.activityKey(node.Name)
				return nil, &InputDiagnostic{Stage: "INPUT_VALIDATION", Code: "INVALID_EXPECTED_VALUE",
					Key: shown, Message: fmt.Sprintf("expected %q: %v", shown, err)}
			}
			expected++
		}
	}
	if roots != len(test.Inputs) || expected != len(test.Expected) || requireExpectation && expected == 0 {
		return nil, fmt.Errorf("inputs and nonempty expectations must name declared activities")
	}
	return append(make([]json.RawMessage, 0, roots), storage[:roots]...), nil
}

func canonicalScalar(raw []byte, entity string) (json.RawMessage, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, fmt.Errorf("explicit %s scalar required", entity)
	}
	var value any
	switch entity {
	case "Integer":
		var integer int64
		if err := json.Unmarshal(raw, &integer); err != nil {
			return nil, fmt.Errorf("Integer requires an exact int64")
		}
		value = integer
	case "Boolean":
		var boolean bool
		if err := json.Unmarshal(raw, &boolean); err != nil {
			return nil, fmt.Errorf("Boolean requires a JSON boolean")
		}
		value = boolean
	case "Text":
		text, err := decodeTextScalar(raw)
		if err != nil {
			return nil, err
		}
		value = text
	default:
		return nil, fmt.Errorf("unsupported scalar entity %q", entity)
	}
	canonical, err := json.Marshal(value)
	return canonical, err
}
