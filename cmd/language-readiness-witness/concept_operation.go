package main

import (
	"os"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
)

func buildConceptOperationObservation(cfg config, conceptArtifact []byte) (readinessartifact.Receipt, error) {
	conceptOperationBinding, err := os.ReadFile(cfg.conceptOperationBinding)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	conceptOperationInputs, err := readConceptOperationInputs(cfg.conceptOperationInputDir)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	return readinessartifact.BuildWithConceptOperationEvidence(
		conceptArtifact, conceptOperationBinding, conceptOperationInputs,
		cfg.root, cfg.conceptOperationScratchDir, cfg.expectedRepository, cfg.expectedSHA,
	)
}
