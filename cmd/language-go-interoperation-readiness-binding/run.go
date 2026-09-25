package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagegointeroperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagegointeroperationbinding"
)

func run(arguments []string) error {
	flags := flag.NewFlagSet("language-go-interoperation-readiness-binding", flag.ContinueOnError)
	root := flags.String("root", ".", "logical repository root")
	head := flags.String("expected-head", "", "exact commit SHA")
	conceptPath := flags.String("concept", "", "language concept artifact")
	readinessPath := flags.String("readiness", "", "complete language readiness artifact")
	interopOutput := flags.String("interop-output", "interop-artifact.json", "Go interoperation artifact output")
	bindingOutput := flags.String("binding-output", "binding-artifact.json", "binding artifact output")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	concept, readiness, err := loadInputs(*root, *head, *conceptPath, *readinessPath)
	if err != nil {
		return err
	}
	interop := languagegointeroperation.Evaluate(languagegointeroperation.Input{ExpectedHeadSHA: *head, ConceptArtifact: concept})
	binding := languagegointeroperationbinding.Evaluate(languagegointeroperationbinding.Input{
		ExpectedHeadSHA: *head, Concept: concept, Readiness: readiness.Snapshot, Interoperation: interop})
	if err := writeOutputs(*interopOutput, *bindingOutput, interop, binding); err != nil {
		return err
	}
	if interop.Decision != languagegointeroperation.DecisionPass || binding.Decision != "PASS" {
		return fmt.Errorf("Go interoperation/binding decisions %s/%s", interop.Decision, binding.Decision)
	}
	return nil
}

func loadInputs(root, head, conceptPath, readinessPath string) (languageconcept.Artifact, readinessartifact.Receipt, error) {
	concept, err := loadJSON[languageconcept.Artifact](conceptPath)
	if err != nil {
		return concept, readinessartifact.Receipt{}, err
	}
	if err := languageconcept.ValidateArtifact(os.DirFS(root), concept); err != nil {
		return concept, readinessartifact.Receipt{}, err
	}
	readiness, err := loadJSON[readinessartifact.Receipt](readinessPath)
	if err != nil {
		return concept, readiness, err
	}
	if err := readinessartifact.Validate(readiness); err != nil {
		return concept, readiness, err
	}
	if readiness.HeadSHA != head {
		return concept, readiness, fmt.Errorf("readiness head %s does not match %s", readiness.HeadSHA, head)
	}
	return concept, readiness, nil
}
