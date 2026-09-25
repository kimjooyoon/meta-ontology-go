package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.executable == "" || cfg.head == "" || cfg.concept == "" || cfg.corpus == "" {
		return fmt.Errorf("root, executable, expected-head, concept, and corpus are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	if err := requireExternal(cfg.root, cfg.executable, target); err != nil {
		return err
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
	report := metaff.Evaluate(metaff.Input{ExpectedHeadSHA: cfg.head, ConceptArtifact: concept,
		RegistryRaw: corpus, Executor: cliruntime.Session{Executable: cfg.executable, Root: cfg.root}})
	if err := writeOrCheck(cfg.output, cfg.check, report); err != nil {
		return err
	}
	if err := metaff.Validate(report, cfg.head); err != nil {
		return err
	}
	if report.Decision != metaff.DecisionPass || report.Resolution != metaff.ResolutionExact {
		return fmt.Errorf("%s/%s: %s", report.Decision, report.Resolution, report.ReasonCode)
	}
	return nil
}
