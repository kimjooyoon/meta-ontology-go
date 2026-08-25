package languagedebug

import (
	"fmt"
	"slices"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != "gooo/language-debug-receipt/v1" || receipt.Digest == "" {
		return fmt.Errorf("debug receipt identity is invalid")
	}
	if seal(receipt).Digest != receipt.Digest || !slices.Equal(receipt.NonClaims, CanonicalNonClaims()) {
		return fmt.Errorf("debug receipt digest or non-claims drifted")
	}
	if receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return fmt.Errorf("debug receipt effects are invalid")
	}
	if receipt.Decision == DecisionPass {
		return validatePaused(receipt)
	}
	if receipt.Decision == DecisionFailClosed && receipt.Reason == "DEBUG_BREAKPOINT_NOT_REACHED" &&
		receipt.Resolution == ResolutionExact && receipt.State == StateRejected && receipt.CurrentEvent == nil {
		return nil
	}
	return fmt.Errorf("debug receipt decision is invalid")
}

func validatePaused(receipt Receipt) error {
	if receipt.Reason != "DEBUG_BREAKPOINT_REACHED" || receipt.Resolution != ResolutionExact ||
		receipt.State != StatePaused || receipt.CurrentEvent == nil || len(receipt.Trace) == 0 ||
		receipt.RemainingEvents < 0 || !validDigest(receipt.ExecutionDigest) {
		return fmt.Errorf("paused debug receipt is invalid")
	}
	last := receipt.Trace[len(receipt.Trace)-1]
	if last != *receipt.CurrentEvent || last.Kind != receipt.Breakpoint {
		return fmt.Errorf("debug breakpoint is not the trace frontier")
	}
	return nil
}
