package adapter

import (
	"fmt"
	"strings"
)

func validateReceiptOutcome(outcome ReceiptOutcome, effect ReceiptWriteEffect) error {
	switch outcome {
	case ReceiptOutcomeAccepted, ReceiptOutcomeRejected, ReceiptOutcomeCancelled, ReceiptOutcomeClosed:
	default:
		return fmt.Errorf("unsupported receipt outcome %q", outcome)
	}
	if effect != ReceiptWriteEffectNone && effect != ReceiptWriteEffectObserved {
		return fmt.Errorf("unsupported receipt write_effect %q", effect)
	}
	if outcome == ReceiptOutcomeRejected || outcome == ReceiptOutcomeCancelled || outcome == ReceiptOutcomeClosed {
		if effect != ReceiptWriteEffectNone {
			return fmt.Errorf("%s receipt must have write_effect=none", outcome)
		}
	}
	return nil
}
func validateReceiptPredecessors(eventRef string, predecessors []ReceiptPredecessor) error {
	seen := make(map[string]struct{}, len(predecessors))
	previous := ""
	for index, predecessor := range predecessors {
		if strings.TrimSpace(predecessor.EventRef) == "" || !validDigest(predecessor.Digest) {
			return fmt.Errorf("predecessor %d is incomplete", index)
		}
		if predecessor.EventRef == eventRef {
			return fmt.Errorf("receipt predecessor replays current event_ref")
		}
		if _, exists := seen[predecessor.EventRef]; exists {
			return fmt.Errorf("receipt predecessor event_ref is duplicated")
		}
		if previous != "" && predecessor.EventRef <= previous {
			return fmt.Errorf("receipt predecessors are not in canonical order")
		}
		seen[predecessor.EventRef] = struct{}{}
		previous = predecessor.EventRef
	}
	return nil
}
