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

func validateInputReport(report inputReport) (string, error) {
	if report.Schema != InputReportSchema {
		return "", fmt.Errorf("report schema: got %q want %q", report.Schema, InputReportSchema)
	}
	if len(report.OperationClaimTransitions) != TransitionTotal {
		return "", fmt.Errorf("transition total: got %d want %d", len(report.OperationClaimTransitions), TransitionTotal)
	}
	seen := make(map[string]struct{}, ClaimTotal)
	for index := 0; index < ClaimTotal; index++ {
		registered := report.OperationClaimTransitions[index]
		resolved := report.OperationClaimTransitions[index+ClaimTotal]
		if registered.Sequence != index+1 || resolved.Sequence != index+ClaimTotal+1 {
			return "", fmt.Errorf("transition sequence mismatch at claim %d", index+1)
		}
		if !isRegisteredEvent(registered.Event) {
			return "", fmt.Errorf("claim %d registration event: %q", index+1, registered.Event)
		}
		if registered.ClaimID == "" || registered.ClaimID != resolved.ClaimID {
			return "", fmt.Errorf("claim %d transition identity mismatch", index+1)
		}
		if _, exists := seen[registered.ClaimID]; exists {
			return "", fmt.Errorf("duplicate claim id %q", registered.ClaimID)
		}
		seen[registered.ClaimID] = struct{}{}
		if registered.TransitionDigest == "" || resolved.TransitionDigest == "" {
			return "", fmt.Errorf("claim %d transition digest missing", index+1)
		}
	}
	accepted := 0
	unavailable := 0
	for _, transition := range report.OperationClaimTransitions[ClaimTotal:] {
		switch {
		case isAcceptedEvent(transition.Event):
			accepted++
		case isUnavailableEvent(transition.Event):
			unavailable++
		default:
			return "", fmt.Errorf("unsupported resolution event %q", transition.Event)
		}
	}
	switch {
	case accepted == ClaimTotal && unavailable == 0:
		return ModeSuccess, nil
	case accepted == 0 && unavailable == ClaimTotal:
		return ModeUnknown, nil
	default:
		return "", fmt.Errorf("mixed claim resolution is outside v1 contract: accepted=%d unavailable=%d", accepted, unavailable)
	}
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

func isRegisteredEvent(event string) bool {
	return event == "CLAIM_REGISTERED" || event == "REGISTERED"
}

func isAcceptedEvent(event string) bool {
	return event == "EVIDENCE_ACCEPTED"
}

func isUnavailableEvent(event string) bool {
	return event == "EVIDENCE_UNAVAILABLE"
}
