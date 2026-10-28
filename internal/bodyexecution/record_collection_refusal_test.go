package bodyexecution

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestRecordCollectionInvalidInputsHaveTypedDiagnosticsBeforeGeneration(t *testing.T) {
	for _, raw := range []string{
		`{"items":null}`, `{"items":[null]}`, `{"items":{}}`,
		`{"items":[{"quantity":1}]}`, `{"items":[{"quantity":1,"price":"7"}]}`,
		`{"items":[{"quantity":1,"quantity":2,"price":7}]}`,
		`{"items":[{"quantity":1,"price":7,"extra":0}]}`,
		`{"items":[{"quantity":1,"price":7}],"extra":0}`,
		`{"items":[],"items":[]}`,
		`{"items":[` + strings.Repeat(`{"quantity":1,"price":7},`, 256) + `{"quantity":1,"price":7}]}`,
	} {
		suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{Inputs: map[string]json.RawMessage{"Total": json.RawMessage(raw)}, Expected: map[string]json.RawMessage{"Total": json.RawMessage(`0`)}}}}
		got, err := GenerateComposition(context.Background(), "bad.gooo", nestedLineItemsSource(t), suite, "")
		diagnostic, inputFailure := errors.AsType[*InputDiagnostic](err)
		if !inputFailure || diagnostic.Stage != "INPUT_VALIDATION" || diagnostic.Code != "INVALID_INPUT_VALUE" ||
			len(got.Steps) != 0 || got.Model != nil {
			t.Fatalf("invalid record input reached generation/model: %s: %v", raw, err)
		}
	}
}
