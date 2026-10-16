package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestFixedValueDeclaredCohortFresh1883(t *testing.T) {
	source, err := os.ReadFile("../../examples/source-owned-cohort-observations/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/source-owned-cohort-observations/evaluation.json")
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Cases []struct {
			Inputs   map[string]json.RawMessage `json:"inputs"`
			Expected map[string]json.RawMessage `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &suite); err != nil || len(suite.Cases) != 6 {
		t.Fatal("six separately authored expectations are required", err)
	}
	for index, row := range suite.Cases {
		input, _ := json.Marshal([]json.RawMessage{row.Inputs["Assess"]})
		value, err := EvaluateFixedValue(context.Background(), "cohort.gooo", source, "Assess", input)
		if err != nil {
			t.Fatal(index, err)
		}
		var actual, expected any
		if json.Unmarshal(value.Value, &actual) != nil || json.Unmarshal(row.Expected["Assess"], &expected) != nil ||
			!reflect.DeepEqual(actual, expected) || value.ActivityID != "urn:cohort:assess" ||
			value.SourceSHA256 != digest(source) || value.ProjectionSHA256 == "" {
			t.Fatalf("source case %d differs: %s", index, value.Value)
		}
	}
}

func TestFixedValueSourceOwnsDenominatorFresh1884(t *testing.T) {
	source, err := os.ReadFile("../../examples/source-owned-cohort-observations/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	changed := []byte(strings.Replace(string(source), "Alpha(seed), Beta(seed), Gamma(seed)", "Alpha(seed), Beta(seed)", 1))
	input := []byte(`[{"ids":["urn:cohort:alpha","urn:cohort:beta","urn:cohort:gamma"]}]`)
	first, err := EvaluateFixedValue(context.Background(), "cohort.gooo", source, "Assess", input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := EvaluateFixedValue(context.Background(), "cohort.gooo", changed, "Assess", input)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Expected int    `json:"expected_count"`
		Bound    int    `json:"bound_count"`
		Unknown  int    `json:"unknown_count"`
		Coverage int    `json:"coverage_bps"`
		Status   string `json:"status"`
	}
	if json.Unmarshal(second.Value, &result) != nil || result.Expected != 2 || result.Bound != 2 ||
		result.Unknown != 1 || result.Coverage != 10000 || result.Status != "FAIL_CLOSED" ||
		first.SourceSHA256 == second.SourceSHA256 || first.ActivityID != second.ActivityID {
		t.Fatal("source declaration did not own the changed cohort", string(second.Value))
	}
	missing := []byte(strings.Replace(string(source), "Gamma(seed)", "Unavailable1884(seed)", 1))
	if _, err := EvaluateFixedValue(context.Background(), "missing.gooo", missing, "Assess", input); err == nil {
		t.Fatal("unknown requested activity was treated as observed data")
	}
}

func TestFixedValueCancellationFresh1885(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, candidate := range []context.Context{nil, ctx} {
		if _, err := EvaluateFixedValue(candidate, "fixed.gooo", []byte("package p"), "Run", []byte("[1]")); err == nil {
			t.Fatal("missing or cancelled context accepted")
		}
	}
}
