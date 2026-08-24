package externalecosystemexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

const referenceEvidencePath = "internal/meta/externalecosystemconformance/evidence/gomacro.json"

func Observe(ctx context.Context, sourceRoot, externalRoot string) (Observation, error) {
	sourceBefore, err := captureRepository(ctx, sourceRoot)
	if err != nil {
		return Observation{}, err
	}
	externalBefore, err := captureRepository(ctx, externalRoot)
	if err != nil {
		return Observation{}, err
	}
	goVersion, err := commandText(ctx, sourceRoot, "go", "env", "GOVERSION")
	if err != nil {
		return Observation{}, err
	}
	moduleGo, err := moduleGoVersion(filepath.Join(externalRoot, "go.mod"))
	if err != nil {
		return Observation{}, err
	}
	runs := []RunObservation{runGoTest(ctx, externalRoot, 1), runGoTest(ctx, externalRoot, 2)}
	externalAfter, err := captureRepository(ctx, externalRoot)
	if err != nil {
		return Observation{}, err
	}
	sourceAfter, err := captureRepository(ctx, sourceRoot)
	if err != nil {
		return Observation{}, err
	}
	return Observation{
		Schema: ObservationSchema, Reference: loadReference(sourceRoot, externalBefore, moduleGo),
		GoVersion: goVersion, Runs: runs, SourceBefore: sourceBefore, SourceAfter: sourceAfter,
		ExternalBefore: externalBefore, ExternalAfter: externalAfter,
	}, nil
}

func loadReference(sourceRoot string, external RepositoryState, moduleGo string) ReferenceReceipt {
	path := filepath.Join(sourceRoot, referenceEvidencePath)
	b, err := os.ReadFile(path)
	available := err == nil
	exact := available && json.Valid(b) && bytes.Contains(b, []byte(ExpectedCommit)) &&
		bytes.Contains(b, []byte(ExpectedTree)) && bytes.Contains(b, []byte(ExpectedModuleGo))
	return ReferenceReceipt{
		Available: available, BindingExact: exact, ContractVersion: ReferenceContractVersion,
		Decision: ExpectedReferenceDecision, Resolution: "EXACT", URL: ExpectedReferenceURL,
		Commit: external.Commit, Tree: external.Tree, ModuleGo: moduleGo,
		EvidencePath: referenceEvidencePath, EvidenceSHA256: Digest(b),
	}
}
