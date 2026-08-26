package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagepackageruntime"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.head == "" || cfg.concept == "" || cfg.corpus == "" {
		return fmt.Errorf("root, expected-head, concept, and corpus are required")
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
	corpus, err := os.ReadFile(cfg.corpus)
	if err != nil {
		return err
	}
	report := languagepackageruntime.Evaluate(languagepackageruntime.Input{
		ExpectedHeadSHA: cfg.head, ConceptArtifact: concept, RegistryRaw: corpus,
	})
	if err := writeOrCheck(cfg.output, cfg.check, report); err != nil {
		return err
	}
	if err := languagepackageruntime.Validate(report, cfg.head); err != nil {
		return err
	}
	if report.Decision != languagepackageruntime.DecisionPass {
		return fmt.Errorf("%s: %s", report.Decision, report.ReasonCode)
	}
	return nil
}
