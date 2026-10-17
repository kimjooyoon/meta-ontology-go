package bodycodegen

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestFixedValueRSSMetadataExample(t *testing.T) {
	source, err := os.ReadFile("../../examples/measurement-metadata/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/measurement-metadata/cases.json")
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
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != 14 {
		t.Fatal("fourteen independent metadata expectations are required")
	}
	const entry = "measurement/rss:ExplainRSSPair"
	for index, row := range suite.Cases {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			input, err := json.Marshal([]json.RawMessage{row.Inputs[entry+".input0"], row.Inputs[entry+".input1"]})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			value, err := EvaluateFixedValue(ctx, "metadata.gooo", source, "ExplainRSSPair", input)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected map[string]string
			if json.Unmarshal(value.Value, &actual) != nil || json.Unmarshal(row.Expected[entry], &expected) != nil ||
				len(expected) != 4 || !maps.Equal(actual, expected) {
				t.Fatalf("declared measurement explanation differs: %s", value.Value)
			}
			if value.ActivityID != "gooo://measurement/explain-rss-pair/v1" ||
				value.SourceSHA256 != digest(source) || value.ProjectionSHA256 == "" {
				t.Fatal("explanation lost its authored source identity")
			}
		})
	}
}
