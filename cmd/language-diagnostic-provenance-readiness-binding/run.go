package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenancebinding"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.head == "" || cfg.concept == "" ||
		cfg.readiness == "" || cfg.provenance == "" {
		return fmt.Errorf("root, expected-head, concept, readiness, and provenance are required")
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
	readiness, err := readJSON[readinessartifact.Receipt](cfg.readiness)
	if err != nil {
		return err
	}
	if err := readinessartifact.Validate(readiness); err != nil {
		return err
	}
	provenance, err := readJSON[languagediagnosticprovenance.Report](cfg.provenance)
	if err != nil {
		return err
	}
	if err := languagediagnosticprovenance.Validate(provenance, cfg.head); err != nil {
		return err
	}
	if readiness.HeadSHA != cfg.head {
		return fmt.Errorf("readiness head %s does not match %s", readiness.HeadSHA, cfg.head)
	}
	binding := languagediagnosticprovenancebinding.Evaluate(
		languagediagnosticprovenancebinding.Input{
			ExpectedHeadSHA: cfg.head, Concept: concept,
			Readiness: readiness.Snapshot, Provenance: provenance,
		},
	)
	if err := writeOrCheck(cfg.output, cfg.check, binding); err != nil {
		return err
	}
	if binding.Decision != "PASS" {
		return fmt.Errorf("diagnostic readiness binding decision %s", binding.Decision)
	}
	return nil
}
