package proofchoicealgebra

import "fmt"

// Combine is the proof-choice algebra's disjoint-union operator (⊕). Exact
// duplicate witnesses are idempotent; conflicting values for one ID are not.
func Combine(left, right Bundle) (Bundle, error) {
	if err := validateBundleShape(left); err != nil {
		return Bundle{}, err
	}
	if err := validateBundleShape(right); err != nil {
		return Bundle{}, err
	}
	result := Bundle{Items: append([]Item(nil), left.Items...), Transitions: append([]Transition(nil), left.Transitions...)}
	for _, item := range right.Items {
		found := false
		for _, existing := range result.Items {
			if existing.ID != item.ID {
				continue
			}
			found = true
			if !sameItem(existing, item) {
				return Bundle{}, fmt.Errorf("PROOF_CHOICE_CONTRADICTION: %s", item.ID)
			}
		}
		if !found {
			result.Items = append(result.Items, item)
		}
	}
	for _, transition := range right.Transitions {
		found := false
		for _, existing := range result.Transitions {
			if existing.ClaimID == transition.ClaimID && existing.From == transition.From && existing.To == transition.To {
				found = true
				if existing.Choice != transition.Choice || !sameTransition(existing, transition) {
					return Bundle{}, fmt.Errorf("PROOF_CHOICE_CONTRADICTION: transition:%s", transition.ClaimID)
				}
			}
		}
		if !found {
			result.Transitions = append(result.Transitions, transition)
		}
	}
	if failure := validateBundle(result); failure != "" {
		return Bundle{}, fmt.Errorf("%s", failure)
	}
	return result, nil
}

func validateBundleShape(bundle Bundle) error {
	for _, item := range bundle.Items {
		if !item.Choice.Valid() {
			return fmt.Errorf("PROOF_CHOICE_MISSING")
		}
	}
	return nil
}

func sameTransition(left, right Transition) bool {
	left.Line, right.Line = 0, 0
	return left == right
}
