package verify

import "sort"

func compose(left, right Evidence) Value {
	value := Value{Contributors: []string{left.Operation, right.Operation}}
	for _, evidence := range []Evidence{left, right} {
		switch deriveState(evidence) {
		case directUnknown:
			value.DirectUnknowns = unique(value.DirectUnknowns, evidence.Operation)
		case dependencyBlocked:
			value.BlockedDependencies = unique(value.BlockedDependencies, evidence.DependencyClaimID)
		case invariantOnly:
			value.PreservedInvariants = unique(value.PreservedInvariants, evidence.InvariantEvidence)
		}
	}
	sort.Strings(value.Contributors)
	sort.Strings(value.DirectUnknowns)
	sort.Strings(value.BlockedDependencies)
	sort.Strings(value.PreservedInvariants)
	switch {
	case len(value.DirectUnknowns) > 0 && len(value.BlockedDependencies) > 0:
		value.State = mixedUnresolved
	case len(value.DirectUnknowns) > 0:
		value.State = directUnknown
	case len(value.BlockedDependencies) > 0:
		value.State = dependencyBlocked
	case len(value.PreservedInvariants) > 0:
		value.State = invariantOnly
	default:
		value.State = exact
	}
	return value
}

func unique(values []string, additions ...string) []string {
	for _, addition := range additions {
		if addition == "" || contains(values, addition) {
			continue
		}
		values = append(values, addition)
	}
	return values
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func classify(value Value) (decision, resolution, reason string, topSuccess bool) {
	switch value.State {
	case exact:
		return "PASS", "EXACT", "ALL_OBSERVATIONS_EXACT", true
	case directUnknown:
		return "UNKNOWN", "LOWER_RESOLUTION", "DIRECT_UNKNOWN_PRESERVED", false
	case dependencyBlocked:
		return "UNKNOWN", "LOWER_RESOLUTION", "DEPENDENCY_BLOCKED_PRESERVED", false
	case invariantOnly:
		return "HOLD", "INVARIANT_ONLY", "KNOWN_INVARIANT_PRESERVED", false
	case mixedUnresolved:
		return "UNKNOWN", "LOWER_RESOLUTION", "MIXED_UNRESOLVED_PRESERVED", false
	default:
		return "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_COMPOSITION_STATE", false
	}
}

func transitionState(state string) string {
	if state == exact {
		return "DISCHARGED"
	}
	return "OPEN"
}
