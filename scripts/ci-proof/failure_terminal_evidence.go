package main

import (
	"fmt"
	"strings"
)

func sameTerminalFailureEvidence(evidence []terminalFailureEvidence, jobs []failureJob, codes []string) bool {
	if len(evidence) == 0 {
		return true
	}
	if len(evidence) != len(jobs) || len(evidence) != len(codes) {
		return false
	}
	for index, item := range evidence {
		if item.Job != jobs[index] || item.Code != codes[index] {
			return false
		}
	}
	return true
}

func validateTerminalFailureEvidence(manifest failureManifest) error {
	if len(manifest.TerminalFailureEvidence) == 0 {
		return nil
	}
	for _, item := range manifest.TerminalFailureEvidence {
		switch item.Classification {
		case "known_job":
			if item.Code == "CI-UNCLASSIFIED-001" || !strings.HasPrefix(item.Reason, "catalog_mapping:") {
				return fmt.Errorf("known terminal failure evidence has an invalid classification reason")
			}
		case "unclassified_job":
			if item.Code != "CI-UNCLASSIFIED-001" || item.Reason != "unknown_terminal_job" {
				return fmt.Errorf("unclassified terminal failure evidence is not fail-closed")
			}
		case "proof_evidence":
			if !strings.HasPrefix(item.Reason, "proof_state:") {
				return fmt.Errorf("proof terminal failure evidence has no exact proof state")
			}
		default:
			return fmt.Errorf("terminal failure evidence classification is unknown")
		}
	}
	return nil
}
