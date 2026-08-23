package main

import (
	"fmt"
	"io"
	"os"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
)

func build(cfg config) (readinessartifact.Receipt, error) {
	raw, err := os.ReadFile(cfg.input)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	if cfg.promotion == "" {
		return readinessartifact.Build(raw, cfg.expectedSHA)
	}
	promotion, err := os.ReadFile(cfg.promotion)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	if cfg.guarded != "" {
		guarded, err := os.ReadFile(cfg.guarded)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
		useCases, err := os.ReadFile(cfg.useCases)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
		syntaxReport, err := os.ReadFile(cfg.syntax)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
		diagnosticReport, err := os.ReadFile(cfg.diagnostic)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
		packageRuntimeReport, err := os.ReadFile(cfg.packageRuntime)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
		toolchainCLIReport, err := os.ReadFile(cfg.toolchainCLI)
		if err != nil {
			return readinessartifact.Receipt{}, err
		}
		return readinessartifact.BuildWithCompleteEvidence(readinessartifact.CompleteEvidenceInput{
			ConceptArtifact: raw, Promotion: promotion, Capability: guarded, UseCases: useCases,
			Syntax: syntaxReport, Diagnostic: diagnosticReport, PackageRuntime: packageRuntimeReport,
			ToolchainCLI: toolchainCLIReport, HeadSHA: cfg.expectedSHA,
		})
	}
	return readinessartifact.BuildWithProposalPromotion(raw, promotion, cfg.expectedSHA)
}

func printSummary(stdout io.Writer, receipt readinessartifact.Receipt) {
	summary := receipt.Snapshot.Summary
	fmt.Fprintf(stdout,
		"language-readiness-artifact: decision=%s completed=%d total=%d bps=%d fixed_point=%s writes=%d digest=%s\n",
		receipt.Snapshot.Decision, summary.Completed, summary.Total, summary.ReadinessBPS,
		receipt.FixedPoint.Decision, receipt.Snapshot.RepositoryWrites, receipt.ArtifactDigest)
}
