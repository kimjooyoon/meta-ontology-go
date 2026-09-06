package generation

import (
	"errors"
	"fmt"
)

// VerifySemanticObservation independently reclassifies a compiler observation
// before a caller is allowed to derive an adoption proposal from it.
func VerifySemanticObservation(observation SemanticObservation) error {
	if err := validateSemanticObservationContract(observation.Contract); err != nil {
		return err
	}
	decision, reason, unknown, err := independentlyClassifySemanticObservation(observation)
	if err != nil {
		return err
	}
	if decision != observation.Decision || reason != observation.Reason {
		return fmt.Errorf("semantic observation decision mismatch: got %s/%s, report %s/%s", decision, reason, observation.Decision, observation.Reason)
	}
	if decision == "UNKNOWN" {
		if !sameEnvelopeUnknown(observation.Unknown, unknown) {
			return errors.New("semantic observation UNKNOWN evidence is not causal")
		}
	} else if observation.Unknown != nil {
		return errors.New("non-unknown semantic observation contains unknown evidence")
	}
	return nil
}

// SemanticObservationUnknownState returns the fixed causal state used when a
// duplicate candidate lacks an exact adoption pair.
func SemanticObservationUnknownState() *EnvelopeUnknownState {
	return semanticObservationUnknownState()
}
