package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func validateArtifact(value artifact) error {
	if value.Schema != artifactSchema || value.Denominator != versionedDenominator {
		return fmt.Errorf("schema/denominator = %q/%d, want %q/%d",
			value.Schema, value.Denominator, artifactSchema, versionedDenominator)
	}
	if value.HeadSHA == "" || value.ContractOperationID != contractOperationID {
		return fmt.Errorf("invalid exact-head or contract operation identity")
	}
	if value.RegistryOperationID != registryOperationID {
		return fmt.Errorf("registry operation = %q, want %q", value.RegistryOperationID, registryOperationID)
	}
	if len(value.RequiredIndicatorIDs) != versionedDenominator {
		return fmt.Errorf("required indicators = %d, want %d", len(value.RequiredIndicatorIDs), versionedDenominator)
	}
	if err := validateScenario(value.Actual, value.RequiredIndicatorIDs); err != nil {
		return fmt.Errorf("actual: %w", err)
	}
	if err := validateScenario(value.MissingEvidence, value.RequiredIndicatorIDs); err != nil {
		return fmt.Errorf("missing evidence: %w", err)
	}
	if value.Actual.Name != productionEvidenceScenario ||
		value.Actual.Decision != decisionPass ||
		value.Actual.PassCount != versionedDenominator ||
		value.Actual.FailCount != 0 || value.Actual.UnknownCount != 0 {
		return fmt.Errorf("actual verdicts = %d/%d/%d %s, want 6/0/0 PASS",
			value.Actual.PassCount, value.Actual.FailCount, value.Actual.UnknownCount, value.Actual.Decision)
	}
	missing := value.MissingEvidence
	if missing.Name != missingEvidenceScenario || missing.Decision != decisionBlock ||
		missing.PassCount != 0 || missing.FailCount != 0 ||
		missing.UnknownCount != versionedDenominator || missing.Resolution != resolutionLower {
		return fmt.Errorf("missing-evidence verdicts = %d/%d/%d %s %s, want 0/0/6 BLOCK LOWER_RESOLUTION",
			missing.PassCount, missing.FailCount, missing.UnknownCount, missing.Decision, missing.Resolution)
	}
	return nil
}

func validateScenario(value scenario, required []string) error {
	if value.Denominator != len(required) {
		return fmt.Errorf("denominator = %d, want %d", value.Denominator, len(required))
	}
	var evaluation transformationeffect.SplitGoEvaluationArtifact
	if err := json.Unmarshal(value.Evaluation, &evaluation); err != nil {
		return err
	}
	if err := transformationeffect.ValidateSplitGoEvaluation(evaluation); err != nil {
		return err
	}
	stats, err := projectEvaluation(value.Evaluation, required)
	if err != nil {
		return err
	}
	if !slices.Equal(
		[]int{value.PassCount, value.FailCount, value.UnknownCount},
		[]int{stats.passCount, stats.failCount, stats.unknownCount},
	) || value.Resolution != stats.resolution {
		return fmt.Errorf("scenario projection does not match embedded evaluation")
	}
	return nil
}
