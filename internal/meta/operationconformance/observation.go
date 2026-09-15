package operationconformance

import "encoding/hex"

func observeIndicator(id string, evidence SplitGoEvidence) Decision {
	if !evidence.EvidenceComplete {
		return DecisionUnknown
	}
	switch id {
	case fixedIndicators[0].ID:
		return observeAtomic(evidence)
	case fixedIndicators[1].ID:
		return observeBuildSemantics(evidence)
	case fixedIndicators[2].ID:
		return observeHeader(evidence)
	case fixedIndicators[3].ID:
		return observeImports(evidence)
	case fixedIndicators[4].ID:
		return observeOrder(evidence)
	case fixedIndicators[5].ID:
		return observePackage(evidence)
	default:
		return DecisionUnknown
	}
}

func observation(definition IndicatorDefinition, decision Decision, evidenceDigest string) IndicatorObservation {
	resolution, value := "EXACT", 0
	if decision == DecisionUnknown {
		resolution = "LOWER_RESOLUTION"
	}
	if decision == DecisionPass {
		value = 1
	}
	return IndicatorObservation{ID: definition.ID, Role: definition.Role, Route: definition.Route,
		RuleID: definition.RuleID, Decision: decision, Resolution: resolution,
		Value: value, Target: 1, ObservationDigest: digestValue([]string{definition.ID, evidenceDigest})}
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
