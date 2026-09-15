package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagedeterministicquery"
)

func run(arguments []string) error {
	flags := flag.NewFlagSet("language-deterministic-query-witness", flag.ContinueOnError)
	root := flags.String("root", ".", "logical repository root")
	head := flags.String("expected-head", "", "exact commit SHA")
	output := flags.String("output", "artifact.json", "query artifact output")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	concept := languageconcept.BuildArtifact(os.DirFS(*root))
	report := languagedeterministicquery.Evaluate(languagedeterministicquery.Input{
		ExpectedHeadSHA: *head, ConceptArtifact: concept,
	})
	if err := writeJSON(*output, report); err != nil {
		return err
	}
	if report.Decision != languagedeterministicquery.DecisionPass {
		return fmt.Errorf("deterministic query decision %s: %s", report.Decision, report.ReasonCode)
	}
	return nil
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
