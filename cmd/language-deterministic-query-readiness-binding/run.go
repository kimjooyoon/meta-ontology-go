package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagedeterministicquery"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagedeterministicquerybinding"
)

func run(arguments []string) error {
	flags := flag.NewFlagSet("language-deterministic-query-readiness-binding", flag.ContinueOnError)
	root := flags.String("root", ".", "logical repository root")
	head := flags.String("expected-head", "", "exact commit SHA")
	queryOutput := flags.String("query-output", "query-artifact.json", "query artifact output")
	bindingOutput := flags.String("binding-output", "binding-artifact.json", "binding artifact output")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	concept := languageconcept.BuildArtifact(os.DirFS(*root))
	conceptRaw, err := json.Marshal(concept)
	if err != nil {
		return err
	}
	readiness, err := languagereadiness.Evaluate(conceptRaw)
	if err != nil {
		return err
	}
	queryReport := languagedeterministicquery.Evaluate(languagedeterministicquery.Input{
		ExpectedHeadSHA: *head, ConceptArtifact: concept,
	})
	binding := languagedeterministicquerybinding.Evaluate(languagedeterministicquerybinding.Input{
		ExpectedHeadSHA: *head, Concept: concept, Readiness: readiness, Query: queryReport,
	})
	if err := writeOutputs(*queryOutput, *bindingOutput, queryReport, binding); err != nil {
		return err
	}
	if queryReport.Decision != languagedeterministicquery.DecisionPass || binding.Decision != "PASS" {
		return fmt.Errorf("query/binding decisions %s/%s", queryReport.Decision, binding.Decision)
	}
	return nil
}

func writeOutputs(queryPath, bindingPath string, queryReport, binding any) error {
	if err := writeJSON(queryPath, queryReport); err != nil {
		return err
	}
	return writeJSON(bindingPath, binding)
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
