package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.head == "" || cfg.concept == "" || cfg.registry == "" {
		return fmt.Errorf("root, expected-head, concept, and registry are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	concept, err := readJSON[languageconcept.Artifact](cfg.concept)
	if err != nil {
		return err
	}
	if err := languageconcept.ValidateArtifact(os.DirFS(cfg.root), concept); err != nil {
		return err
	}
	registry, err := readJSON[languagediagnosticprovenance.CaseRegistry](cfg.registry)
	if err != nil {
		return err
	}
	report := languagediagnosticprovenance.Evaluate(
		languagediagnosticprovenance.Input{
			ExpectedHeadSHA: cfg.head, ConceptArtifact: concept, Registry: registry,
		},
	)
	if err := writeOrCheck(cfg.output, cfg.check, report); err != nil {
		return err
	}
	if err := languagediagnosticprovenance.Validate(report, cfg.head); err != nil {
		return fmt.Errorf("%w: reason=%s summary=%+v",
			err, report.ReasonCode, report.Summary)
	}
	return nil
}
