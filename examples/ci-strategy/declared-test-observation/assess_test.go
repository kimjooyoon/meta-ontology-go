package declaredtest

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func equalJSON(t *testing.T, actual, expected []byte) {
	t.Helper()
	var got, want any
	if err := json.Unmarshal(actual, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(expected, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("assessment differs: got=%s want=%s", actual, expected)
	}
}

func TestAuthoredAssessmentsFromGoooSource(t *testing.T) {
	raw, err := os.ReadFile("cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases struct {
		Cases []struct{ Inputs, Expected map[string]json.RawMessage }
	}
	if err := json.Unmarshal(raw, &cases); err != nil || len(cases.Cases) != 21 {
		t.Fatal("independent authored denominator changed", err, len(cases.Cases))
	}
	labels, err := os.ReadFile("case-labels.json")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(labels, &names); err != nil || len(names) != len(cases.Cases) {
		t.Fatal("case labels lost their original positions", err)
	}
	program, err := bodycodegen.NewFixedProgram(context.Background(), sourcePath, source, "Assess")
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range cases.Cases {
		t.Run(names[i], func(t *testing.T) {
			input, err := json.Marshal([]json.RawMessage{row.Inputs["Assess"]})
			if err != nil {
				t.Fatal(err)
			}
			actual, err := program.Evaluate(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, actual.Value, row.Expected["Assess"])
		})
	}
}
