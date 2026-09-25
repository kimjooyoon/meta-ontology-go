package causality

import (
	"encoding/json"
	"fmt"
)

func parseInputReport(input []byte, expectedMode string) (inputReport, string, error) {
	var report inputReport
	if err := json.Unmarshal(input, &report); err != nil {
		return inputReport{}, "", fmt.Errorf("decode operation catalog report: %w", err)
	}
	mode, err := validateInputReport(report)
	if err != nil {
		return inputReport{}, "", err
	}
	if expectedMode != "" && expectedMode != mode {
		return inputReport{}, "", fmt.Errorf("report mode: got %q want %q", mode, expectedMode)
	}
	return report, mode, nil
}

func (report inputReport) transitionHead() string {
	if report.OperationClaimTransitionHead != "" {
		return report.OperationClaimTransitionHead
	}
	return report.OperationClaimTransitionHeadDigest
}

func bindingEvidence(sourceDigest, semanticIRDigest string) []string {
	missing := make([]string, 0, 3)
	if sourceDigest == "" {
		missing = append(missing, "source_digest")
	}
	if semanticIRDigest == "" {
		missing = append(missing, "semantic_ir_digest")
	}
	return append(missing, "source_ir_binding_digest")
}
