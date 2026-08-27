package operationprovenance

import (
	"fmt"
	"path/filepath"
)

func Build(source []byte, observation WorkspaceObservation) (Receipt, error) {
	return BuildWithArtifacts(source, filepath.Join("examples", "meta-operation-provenance", "artifacts"), observation)
}

func BuildWithArtifacts(source []byte, artifactRoot string, observation WorkspaceObservation) (Receipt, error) {
	ir, err := lowerSource(source)
	if err != nil {
		return Receipt{}, err
	}
	metrics, scenarios, reconstruction, sourceIssues, err := reconstructSemanticData(ir)
	if err != nil {
		return Receipt{}, err
	}
	families, err := validateContract(metrics, scenarios)
	if err != nil {
		return Receipt{}, err
	}
	artifacts, err := collectArtifacts(artifactRoot, metrics)
	if err != nil {
		return Receipt{}, err
	}
	receipt := Receipt{
		Schema: ReceiptSchema, Toolchain: Toolchain,
		SourceDigest:            digestBytes(source),
		CanonicalSemanticDigest: "sha256:" + ir.StableHash(),
		SourceResolution:        sourceResolution(sourceIssues), SourceReconstruction: reconstruction, SourceIssues: sourceIssues, FamilyCardinality: families, WorkspaceObservation: observation,
		Scenarios: make([]ScenarioResult, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		result, err := evaluateScenario(metrics, scenario, artifacts, receipt.SourceDigest, receipt.CanonicalSemanticDigest)
		if err != nil {
			return Receipt{}, err
		}
		result.SourceResolution = receipt.SourceResolution
		receipt.Scenarios = append(receipt.Scenarios, result)
	}
	return sealReceipt(receipt)
}

// BuildObserved binds the receipt to isolated repository before/after status.
func BuildObserved(source []byte, repositoryRoot string) (Receipt, error) {
	artifactRoot := filepath.Join(repositoryRoot, "examples", "meta-operation-provenance", "artifacts")
	return BuildObservedWithArtifacts(source, artifactRoot, repositoryRoot)
}

func BuildObservedWithArtifacts(source []byte, artifactRoot, repositoryRoot string) (Receipt, error) {
	before, err := readRepositorySnapshot(repositoryRoot)
	if err != nil {
		return Receipt{}, fmt.Errorf("observe repository before producer: %w", err)
	}
	receipt, buildErr := BuildWithArtifacts(source, artifactRoot, WorkspaceObservation{})
	after, afterErr := readRepositorySnapshot(repositoryRoot)
	if buildErr != nil {
		return Receipt{}, buildErr
	}
	if afterErr != nil {
		return Receipt{}, fmt.Errorf("observe repository after producer: %w", afterErr)
	}
	receipt.WorkspaceObservation = deriveObservation(before, after)
	return sealReceipt(receipt)
}
