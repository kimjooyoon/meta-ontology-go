package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestFixedValueOpenFGAModelChangeExample(t *testing.T) {
	source, err := os.ReadFile("../../examples/openfga-model-change/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/openfga-model-change/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Schema string `json:"schema"`
		Cases  []struct {
			Inputs   map[string]json.RawMessage `json:"inputs"`
			Expected map[string]json.RawMessage `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		t.Fatal(err)
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != 21 {
		t.Fatal("the retained comparison corpus must contain 21 unique input cases")
	}
	for index, row := range suite.Cases {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			input, err := json.Marshal([]json.RawMessage{row.Inputs["Compare"]})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			value, err := EvaluateFixedValue(ctx, "comparison.gooo", source, "Compare", input)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected map[string]any
			if json.Unmarshal(value.Value, &actual) != nil || json.Unmarshal(row.Expected["Compare"], &expected) != nil ||
				len(expected) != 15 || !reflect.DeepEqual(actual, expected) {
				t.Fatalf("comparison differs from its authored result: %s", value.Value)
			}
			if value.ActivityID != "gooo://authorization-delta/compare" ||
				value.SourceSHA256 != digest(source) || value.ProjectionSHA256 == "" {
				t.Fatal("comparison lost its authored source identity")
			}
		})
	}
}
