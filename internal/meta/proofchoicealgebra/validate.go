package proofchoicealgebra

func validateBundle(bundle Bundle) string {
	if len(bundle.Items) == 0 {
		return "NO_PROOF_CHOICES"
	}
	byID, claims, failure := validateItems(bundle.Items)
	if failure != "" {
		return failure
	}
	if failure := validateTransitions(bundle.Transitions, byID, claims); failure != "" {
		return failure
	}
	return missingPersistentTransitions(bundle.Transitions, claims)
}

func missingPersistentTransitions(transitions []Transition, claims map[string]Item) string {
	present := make(map[string]bool, len(transitions))
	for _, transition := range transitions {
		present[transition.ClaimID] = true
	}
	for id := range claims {
		if !present[id] {
			return "PERSISTENT_TRANSITION_MISSING"
		}
	}
	return ""
}
