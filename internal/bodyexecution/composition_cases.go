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
	Schema string            `json:"schema"`
	Cases  []CompositionCase `json:"cases"`
}

func DecodeCompositionCases(raw []byte) (CompositionCases, error) {
	var suite CompositionCases
	if err := decode(raw, &suite, 32<<10); err != nil {
		return suite, err
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) < 1 || len(suite.Cases) > 128 {
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
	if len(suite.Cases) < 1 || len(suite.Cases) > 128 {
		return nil, fmt.Errorf("composition cases require schema and 1..128 cases")
	}
	raw, err := json.Marshal(suite)
	if err != nil || len(raw) > 32<<10 {
		return nil, fmt.Errorf("composition cases exceed 32 KiB")
	}
	rows := make([][]json.RawMessage, len(suite.Cases))
	for c, test := range suite.Cases {
		row, err := graph.inputRow(test, suite.Schema != CompositionInputsSchema)
		if err != nil {
			return nil, fmt.Errorf("case %d: %w", c, err)
		}
		rows[c] = row
	}
	return rows, nil
}

func (graph compositionGraph) inputRow(test CompositionCase, requireExpectation bool) ([]json.RawMessage, error) {
	var storage [compositionLimit * compositionLimit]json.RawMessage
	roots, expected := 0, 0
	for _, node := range graph.nodes[:graph.count] {
		for _, input := range node.inputSlots() {
			key := node.inputKey(input.Port)
			value, hasInput := test.Inputs[key]
			if input.From < 0 {
				if !hasInput {
					return nil, fmt.Errorf("root input %q requires an explicit value", key)
				}
				canonical, err := graph.canonicalValue(value, input.Type)
				if err != nil {
					return nil, fmt.Errorf("input %q: %w", key, err)
				}
				storage[roots], roots = canonical, roots+1
			} else if hasInput {
				return nil, fmt.Errorf("bound input %q must receive its producer's result", key)
			}
		}
		if value, present := test.Expected[node.Name]; present {
			if _, err := graph.canonicalValue(value, node.OutputType); err != nil {
				return nil, fmt.Errorf("expected %q: %w", node.Name, err)
			}
			expected++
		}
	}
	if roots != len(test.Inputs) || expected != len(test.Expected) || requireExpectation && expected == 0 {
		return nil, fmt.Errorf("inputs and nonempty expectations must name declared activities")
	}
	return append([]json.RawMessage(nil), storage[:roots]...), nil
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
		var text string
		if err := json.Unmarshal(raw, &text); err != nil || len(text) > 1024 {
			return nil, fmt.Errorf("Text requires a string of at most 1024 UTF-8 bytes")
		}
		value = text
	default:
		return nil, fmt.Errorf("unsupported scalar entity %q", entity)
	}
	canonical, err := json.Marshal(value)
	return canonical, err
}
