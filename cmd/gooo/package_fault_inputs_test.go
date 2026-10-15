package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestFaultDiagnosticInputExampleMatchesSavedInputs(t *testing.T) {
	raw, err := os.ReadFile("../../examples/compound-updates/division.inputs.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionInputs(raw)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile("../../docs/observations/local-division-origins/first-observations.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Cases bodyexecution.CompositionCases
	}
	if err := json.Unmarshal(readNamedInputArchive(t, archive)["public-outputs/public-local-division.json"], &saved); err != nil {
		t.Fatal(err)
	}
	if len(suite.Cases) != len(saved.Cases.Cases) {
		t.Fatal("input count differs")
	}
	for i, row := range suite.Cases {
		if len(row.Inputs) != 2 || len(row.Expected) != 0 {
			t.Fatal("input-only example changed", i)
		}
		for _, port := range []string{"input0", "input1"} {
			if string(row.Inputs["examples/compound:Divide."+port]) != string(saved.Cases.Cases[i].Inputs["Divide."+port]) {
				t.Fatal("saved input value differs", i, port)
			}
		}
	}
}
