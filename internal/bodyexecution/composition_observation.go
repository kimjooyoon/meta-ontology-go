package bodyexecution

import (
	"encoding/json"
	"fmt"
)

const CompositionInputsSchema = "gooo/body-composition-inputs/v1"

// DecodeCompositionInputs requests actual values without supplying an oracle.
// The distinct schema keeps zero expectations separate from a finite test suite.
func DecodeCompositionInputs(raw []byte) (CompositionCases, error) {
	var input struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}
	if err := decode(raw, &input, 32<<10); err != nil {
		return CompositionCases{}, err
	}
	if input.Schema != CompositionInputsSchema || len(input.Inputs) < 1 || len(input.Inputs) > 128 {
		return CompositionCases{}, fmt.Errorf("composition inputs require schema and 1..128 input rows")
	}
	suite := CompositionCases{Schema: CompositionInputsSchema, Cases: make([]CompositionCase, len(input.Inputs))}
	for i, values := range input.Inputs {
		suite.Cases[i].Inputs = values
	}
	return suite, nil
}

func validateCompositionInputMode(suite CompositionCases) error {
	if suite.Schema == CompositionInputsSchema {
		for _, row := range suite.Cases {
			if len(row.Expected) != 0 {
				return fmt.Errorf("input-only observations cannot contain expected values")
			}
		}
		return nil
	}
	if suite.Schema != "gooo/body-composition-cases/v1" {
		return fmt.Errorf("unsupported composition input schema")
	}
	return nil
}
