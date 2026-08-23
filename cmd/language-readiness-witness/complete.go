package main

import (
	"os"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
)

func buildComplete(cfg config, concept, promotion []byte) (readinessartifact.Receipt, error) {
	paths := []string{cfg.guarded, cfg.useCases, cfg.syntax, cfg.diagnostic,
		cfg.packageRuntime, cfg.toolchainCLI, cfg.toolchainFormatFix}
	evidence, err := readCompleteEvidence(paths)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	input := readinessartifact.CompleteEvidenceInput{
		ConceptArtifact: concept, Promotion: promotion, Capability: evidence[0],
		UseCases: evidence[1], Syntax: evidence[2], Diagnostic: evidence[3],
		PackageRuntime: evidence[4], ToolchainCLI: evidence[5],
		ToolchainFormatFix: evidence[6], HeadSHA: cfg.expectedSHA,
	}
	if cfg.toolchainConformance != "" {
		input.ToolchainConformance, err = os.ReadFile(cfg.toolchainConformance)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
	}
	if cfg.toolchainLSP != "" {
		input.ToolchainLSP, err = os.ReadFile(cfg.toolchainLSP)
		if err != nil { return readinessartifact.Receipt{}, err }
	}
	return readinessartifact.BuildWithCompleteEvidence(input)
}

func readCompleteEvidence(paths []string) ([][]byte, error) {
	evidence := make([][]byte, len(paths))
	for index, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		evidence[index] = raw
	}
	return evidence, nil
}
