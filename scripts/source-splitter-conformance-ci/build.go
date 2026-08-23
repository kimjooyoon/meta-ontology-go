package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type evidenceEnvelope struct {
	OperationID     string `json:"operation_id"`
	ExpectedHeadSHA string `json:"expected_head_sha"`
}

func buildArtifact(opts options) (artifact, error) {
	contractRaw, err := os.ReadFile(opts.contract)
	if err != nil {
		return artifact{}, fmt.Errorf("read contract: %w", err)
	}
	evidenceRaw, err := os.ReadFile(opts.evidence)
	if err != nil {
		return artifact{}, fmt.Errorf("read evidence: %w", err)
	}
	if err := validateEnvelope(evidenceRaw, opts.expectedSHA); err != nil {
		return artifact{}, err
	}
	binding, err := splitGoBinding()
	if err != nil {
		return artifact{}, err
	}
	required := append([]string(nil), binding.RequiredIndicatorIDs...)
	actual, err := evaluateScenario(
		productionEvidenceScenario, contractRaw, evidenceRaw, required, string(binding.ProofChoice),
	)
	if err != nil {
		return artifact{}, err
	}
	missing, err := evaluateScenario(
		missingEvidenceScenario, contractRaw, []byte("{}"), required, string(binding.ProofChoice),
	)
	if err != nil {
		return artifact{}, err
	}
	return artifact{
		Schema: artifactSchema, HeadSHA: opts.expectedSHA,
		ContractOperationID: contractOperationID,
		RegistryOperationID: registryOperationID,
		RequiredIndicatorIDs: required, Denominator: len(required),
		Actual: actual, MissingEvidence: missing,
	}, nil
}

func validateEnvelope(raw []byte, expectedSHA string) error {
	var envelope evidenceEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("decode evidence envelope: %w", err)
	}
	if envelope.OperationID != contractOperationID {
		return fmt.Errorf("operation_id = %q, want %q", envelope.OperationID, contractOperationID)
	}
	if envelope.ExpectedHeadSHA != expectedSHA {
		return fmt.Errorf("expected_head_sha = %q, want %q", envelope.ExpectedHeadSHA, expectedSHA)
	}
	return nil
}
