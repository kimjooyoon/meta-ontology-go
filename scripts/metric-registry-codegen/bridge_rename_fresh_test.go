package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestSourceAndTargetRenameFresh1703(t *testing.T) {
	directory := "../../internal/meta/metricprogram"
	raw, err := os.ReadFile(filepath.Join(directory, "registry.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	renamed := strings.NewReplacer(
		"entity CohortTerminalDeclaration id", "entity TerminalRecord id",
		"-> CohortTerminalDeclaration", "-> TerminalRecord",
		"return CohortTerminalDeclaration{", "return TerminalRecord{",
		"field Subject id", "field ExpectedSubject id", "Subject:", "ExpectedSubject:",
	).Replace(string(raw))
	fixture := t.TempDir()
	writeBridgeFixtureFresh1699(t, filepath.Join(fixture, "registry.gooo"), renamed)
	for _, name := range []string{"types_program.go", "source.go", "cohort.go"} {
		content, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "cohort.go" {
			content = []byte(strings.Replace(string(content), "\tSubject", "\tChosenSubject", 1))
		}
		writeBridgeFixtureFresh1699(t, filepath.Join(fixture, name), string(content))
	}
	output := filepath.Join(t.TempDir(), "registry.generated.go")
	if err := generate(filepath.Join(fixture, "registry.gooo"), output); err != nil {
		t.Fatal(err)
	}
	projection, err := bodycodegen.Generate("renamed.gooo", []byte(renamed), "CohortTerminalAt")
	if err != nil {
		t.Fatal(err)
	}
	var expectedField string
	for _, record := range projection.Report.RecordTypes {
		if record.ID == "gooo://metric-registry/CohortTerminalDeclaration" {
			for _, field := range record.Fields {
				if field.ID == "gooo://metric-registry/CohortTerminalDeclaration/Subject" && field.Name == "ExpectedSubject" {
					expectedField = field.GoName
				}
			}
		}
	}
	var combined strings.Builder
	for name, content := range projectionFiles(t, filepath.Dir(output)) {
		if filepath.Ext(name) == ".go" {
			combined.Write(content)
		}
	}
	if expectedField == "" || !strings.Contains(combined.String(), "ChosenSubject: row."+expectedField) {
		t.Fatal("Gooo and native display renames broke the stable field connection")
	}
}
