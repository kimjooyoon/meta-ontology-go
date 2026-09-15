package main

import (
	"flag"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagedeterministicquery"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagedeterministicquerybinding"
	"os"
)

func run(arguments []string) error {
	flags := flag.NewFlagSet("language-deterministic-query-readiness-binding", flag.ContinueOnError)
	root := flags.String("root", ".", "logical repository root")
	head := flags.String("expected-head", "", "exact commit SHA")
	conceptPath := flags.String("concept", "", "language concept artifact")
	readinessPath := flags.String("readiness", "", "complete language readiness artifact")
	queryOutput := flags.String("query-output", "query-artifact.json", "query artifact output")
	bindingOutput := flags.String("binding-output", "binding-artifact.json", "binding artifact output")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	concept, err := loadJSON[languageconcept.Artifact](*conceptPath)
	if err != nil {
		return err
	}
	if err := languageconcept.ValidateArtifact(os.DirFS(*root), concept); err != nil {
		return err
	}
	readiness, err := loadJSON[readinessartifact.Receipt](*readinessPath)
	if err != nil {
		return err
	}
	if err := readinessartifact.Validate(readiness); err != nil {
		return err
	}
	if readiness.HeadSHA != *head {
		return fmt.Errorf("readiness head %s does not match %s", readiness.HeadSHA, *head)
	}
	queryReport := languagedeterministicquery.Evaluate(languagedeterministicquery.Input{
		ExpectedHeadSHA: *head, ConceptArtifact: concept,
	})
	binding := languagedeterministicquerybinding.Evaluate(languagedeterministicquerybinding.Input{
		ExpectedHeadSHA: *head, Concept: concept, Readiness: readiness.Snapshot, Query: queryReport,
	})
	if err := writeOutputs(*queryOutput, *bindingOutput, queryReport, binding); err != nil {
		return err
	}
	if queryReport.Decision != languagedeterministicquery.DecisionPass || binding.Decision != "PASS" {
		return fmt.Errorf("query/binding decisions %s/%s", queryReport.Decision, binding.Decision)
	}
	return nil
}
