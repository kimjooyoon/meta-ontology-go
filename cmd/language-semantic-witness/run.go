package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic"
	"os"
	"path/filepath"
)

func run(arguments []string) error {
	flags := flag.NewFlagSet("language-semantic-witness", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	root := flags.String("root", "", "repository root")
	head := flags.String("head", "", "exact 40-character subject SHA")
	registry := flags.String("registry", "", "versioned semantic corpus")
	syntaxArtifact := flags.String("syntax-artifact", "", "exact-head language syntax receipt")
	output := flags.String("output", "", "write a new receipt outside the repository")
	check := flags.String("check", "", "compare against an existing receipt")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 || *root == "" || *head == "" || *registry == "" || *syntaxArtifact == "" {
		return fmt.Errorf("usage: language-semantic-witness --root ROOT --head SHA --registry FILE --syntax-artifact FILE (--output FILE | --check FILE)")
	}
	if (*output == "") == (*check == "") {
		return fmt.Errorf("exactly one of --output or --check is required")
	}
	rootAbs, err := filepath.Abs(*root)
	if err != nil {
		return err
	}
	registryPath := resolve(rootAbs, *registry)
	syntaxPath := resolve(rootAbs, *syntaxArtifact)
	report, err := languagesemantic.Evaluate(languagesemantic.Input{
		Root:               rootAbs,
		ExpectedHeadSHA:    *head,
		RegistryPath:       registryPath,
		SyntaxArtifactPath: syntaxPath,
	})
	if err != nil {
		return err
	}
	raw, err := marshal(report)
	if err != nil {
		return err
	}
	if *check != "" {
		expected, err := os.ReadFile(*check)
		if err != nil {
			return err
		}
		if !bytes.Equal(expected, raw) {
			return fmt.Errorf("semantic receipt replay differs from %s", *check)
		}
	} else {
		if err := writeOutsideRepository(rootAbs, *output, raw); err != nil {
			return err
		}
	}
	fmt.Printf("language-semantic: decision=%s resolution=%s satisfied=%d/%d source_models=%d laws=%d/%d/%d writes=%d\n",
		report.Decision, report.Resolution, report.Summary.Satisfied, report.Summary.Total, report.Summary.SourceModels,
		report.Summary.PresentationLaws, report.Summary.CandidateAuthorityLaws, report.Summary.DeterministicAuthorityLaws, report.RepositoryWrites)
	if report.Decision != languagesemantic.DecisionPass || report.Resolution != languagesemantic.ResolutionExact {
		return fmt.Errorf("semantic model failed closed: %s", report.ReasonCode)
	}
	return nil
}
