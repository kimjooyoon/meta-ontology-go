package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	metaconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.head == "" || cfg.concept == "" || cfg.corpus == "" {
		return fmt.Errorf("root, expected-head, concept, and corpus are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	paths := cfg.artifactPaths()
	external := []string{cfg.concept, target}
	for _, path := range paths {
		if path == "" {
			return fmt.Errorf("all nine conformance receipts are required")
		}
		external = append(external, path)
	}
	if err := requireExternal(cfg.root, external...); err != nil {
		return err
	}
	conceptRaw, err := readFile(cfg.concept)
	if err != nil {
		return err
	}
	concept := languageconcept.Artifact{}
	if err := decodeJSON(conceptRaw, &concept); err != nil {
		return err
	}
	if err := languageconcept.ValidateArtifact(osDirFS(cfg.root), concept); err != nil {
		return err
	}
	registry, err := readFile(cfg.corpus)
	if err != nil {
		return err
	}
	artifacts, err := readArtifacts(paths)
	if err != nil {
		return err
	}
	report := metaconformance.Evaluate(metaconformance.Input{
		ExpectedHeadSHA: cfg.head, ConceptArtifact: conceptRaw,
		RegistryRaw: registry, Artifacts: artifacts,
	})
	if err := metaconformance.Validate(report, cfg.head); err != nil {
		return err
	}
	return writeOrCheck(cfg.output, cfg.check, report)
}
