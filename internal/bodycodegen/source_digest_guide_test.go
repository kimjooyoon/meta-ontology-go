package bodycodegen

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestFixedSourceDigestGuideRetainsExactPathsAndUncertainty(t *testing.T) {
	source, err := os.ReadFile("../../examples/source-digest-guide/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/source-digest-guide/cases.json")
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
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != 12 {
		t.Fatal("guide requires eight known paths and four unresolved queries")
	}
	program, err := NewFixedProgram(t.Context(), "digest-guide.gooo", source, "Explain")
	if err != nil {
		t.Fatal(err)
	}
	const entry = "example/source-digest-guide:Explain"
	for index, row := range suite.Cases {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			input, err := json.Marshal([]json.RawMessage{row.Inputs[entry]})
			if err != nil {
				t.Fatal(err)
			}
			value, err := program.Evaluate(t.Context(), input)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected map[string]any
			if json.Unmarshal(value.Value, &actual) != nil || json.Unmarshal(row.Expected[entry], &expected) != nil ||
				len(expected) != 6 || !reflect.DeepEqual(actual, expected) {
				t.Fatalf("guide differs from the declared query result: %s", value.Value)
			}
			if value.ActivityID != "urn:gooo:source-digest-guide:explain" ||
				value.SourceSHA256 != digest(source) || value.ProjectionSHA256 == "" {
				t.Fatal("guide lost its original Gooo source or activity identity")
			}
		})
	}
}
