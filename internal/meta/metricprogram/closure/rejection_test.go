package closure_test

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"
)

func TestBuildRejectsTamperedSource(t *testing.T) {
	input := fixtureInput()
	input.Source = append(input.Source, []byte("entity Tampered\n")...)
	if _, err := closure.Build(input); err == nil {
		t.Fatal("expected source digest rejection")
	}
}

func TestBuildRejectsRequiredRootREADME(t *testing.T) {
	input := fixtureInput()
	var program map[string]any
	if err := json.Unmarshal(input.ProgramJSON, &program); err != nil {
		t.Fatal(err)
	}
	root := program["root_policy"].(map[string]any)
	root["readme_requirement"] = "REQUIRED"
	input.ProgramJSON, _ = json.Marshal(program)
	if _, err := closure.Build(input); err == nil {
		t.Fatal("expected project-root README rejection")
	}
}
