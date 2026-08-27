package proofchoicealgebra

import "strings"

func validateTransitions(transitions []Transition, byID map[string]Item, claims map[string]Item) string {
	for _, transition := range transitions {
		claim, exists := claims[transition.ClaimID]
		if !exists || byID[transition.ClaimID].Kind != Claim || !transition.Persistent || transition.From == "" || transition.To == "" {
			return "PERSISTENT_TRANSITION_MISMATCH"
		}
		if transition.Choice != claim.Choice || !transition.Choice.Valid() {
			return "PROOF_CHOICE_CONTRADICTION"
		}
		if metadataUnknown(transition.Producer, transition.Consumer, transition.MetaOperation, transition.Stage, transition.Step, transition.Reason) || strings.EqualFold(transition.From, "UNKNOWN") || strings.EqualFold(transition.To, "UNKNOWN") {
			return "UNKNOWN_CONTEXT"
		}
	}
	return ""
}
