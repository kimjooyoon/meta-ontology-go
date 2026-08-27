package proofchoicejudge

import "strings"

func validateTransitions(transitions []transition, claims map[string]string) string {
	seen := map[string]bool{}
	for _, value := range transitions {
		choice, exists := claims[value.ClaimID]
		if !exists || !value.Persistent || value.From == "" || value.To == "" {
			return "PERSISTENT_TRANSITION_MISMATCH"
		}
		if !validChoice(value.Choice) || choice != value.Choice {
			return "PROOF_CHOICE_CONTRADICTION"
		}
		if unknown(value.Producer, value.Consumer, value.MetaOperation, value.Stage, value.Step, value.Reason) || unknown(value.From, value.To) {
			return "UNKNOWN_CONTEXT"
		}
		seen[value.ClaimID] = true
	}
	for claimID := range claims {
		if !seen[claimID] {
			return "PERSISTENT_TRANSITION_MISSING"
		}
	}
	return ""
}

func sameItem(left, right item) bool {
	left.Line, right.Line = 0, 0
	return left == right
}

func unknown(values ...string) bool {
	for _, value := range values {
		if value == "" || strings.EqualFold(value, "UNKNOWN") {
			return true
		}
	}
	return false
}
