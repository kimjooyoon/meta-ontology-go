package main

import (
	"os"
	"path/filepath"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func buildComplete(cfg config, concept, promotion []byte) (readinessartifact.Receipt, error) {
	paths := []string{cfg.guarded, cfg.useCases, cfg.syntax, cfg.diagnostic,
		cfg.packageRuntime, cfg.toolchainCLI, cfg.toolchainFormatFix}
	evidence, err := readCompleteEvidence(paths)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	conceptOperationBinding, err := os.ReadFile(cfg.conceptOperationBinding)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	conceptOperationInputs, err := readConceptOperationInputs(cfg.conceptOperationInputDir)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	input := readinessartifact.CompleteEvidenceInput{
		ConceptArtifact: concept, ConceptOperationBinding: conceptOperationBinding, Promotion: promotion, Capability: evidence[0],
		ConceptOperationInputs: conceptOperationInputs, RepositoryRoot: cfg.root,
		UseCases: evidence[1], Syntax: evidence[2], Diagnostic: evidence[3],
		PackageRuntime: evidence[4], ToolchainCLI: evidence[5],
		ToolchainFormatFix: evidence[6], ExpectedRepository: cfg.expectedRepository,
		HeadSHA: cfg.expectedSHA, ExpectedPredecessorSHA: cfg.expectedPredecessorSHA,
	}
	if cfg.toolchainConformance != "" {
		input.ToolchainConformance, err = os.ReadFile(cfg.toolchainConformance)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
	}
	if cfg.toolchainLSP != "" {
		input.ToolchainLSP, err = os.ReadFile(cfg.toolchainLSP)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
	}
	if cfg.toolchainRelease != "" {
		input.ToolchainRelease, err = os.ReadFile(cfg.toolchainRelease)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
	}
	return readinessartifact.BuildWithCompleteEvidence(input)
}

func readConceptOperationInputs(directory string) (conceptoperation.SourceInputs, error) {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(directory, name))
	}
	strategy, err := read("strategy-plan.json")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	strategyVerification, err := read("strategy-verification.json")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	sourceMetrics, err := read("source-metrics.json")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	intervention, err := read("intervention-ledger.json")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	interventionVerification, err := read("intervention-verification.json")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	program, err := read("program.json")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	programSource, err := read("program.gooo")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	programVerification, err := read("verification.json")
	if err != nil {
		return conceptoperation.SourceInputs{}, err
	}
	return conceptoperation.SourceInputs{
		Strategy: strategy, StrategyVerification: strategyVerification, SourceMetrics: sourceMetrics,
		Intervention: intervention, InterventionVerification: interventionVerification,
		Program: program, ProgramSource: programSource, ProgramVerification: programVerification,
	}, nil
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
