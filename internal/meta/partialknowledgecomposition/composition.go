package partialknowledgecomposition

import "slices"

// Compose combines source-derived evidence without manufacturing a more
// precise claim. Its output retains every unresolved cause and every known
// invariant.
func Compose(left, right Evidence) Value {
	value := Value{Contributors: []string{left.Operation, right.Operation}}
	for _, evidence := range []Evidence{left, right} {
		switch deriveState(evidence) {
		case StateDirectUnknown:
			value.DirectUnknowns = appendUnique(value.DirectUnknowns, evidence.Operation)
		case StateDependencyBlocked:
			if evidence.Dependency != nil {
				value.BlockedDependencies = appendUnique(value.BlockedDependencies, evidence.Dependency.ClaimID)
			}
		case StateInvariantOnly:
			value.PreservedInvariants = appendUnique(value.PreservedInvariants, evidence.InvariantEvidence)
		}
	}
	value.DirectUnknowns = sortedUnique(value.DirectUnknowns)
	value.BlockedDependencies = sortedUnique(value.BlockedDependencies)
	value.PreservedInvariants = sortedUnique(value.PreservedInvariants)
	slices.Sort(value.Contributors)
	switch {
	case len(value.DirectUnknowns) > 0 && len(value.BlockedDependencies) > 0:
		value.State = StateMixedUnresolved
	case len(value.DirectUnknowns) > 0:
		value.State = StateDirectUnknown
	case len(value.BlockedDependencies) > 0:
		value.State = StateDependencyBlocked
	case len(value.PreservedInvariants) > 0:
		value.State = StateInvariantOnly
	default:
		value.State = StateExact
	}
	return value
}

func appendUnique(values []string, additions ...string) []string {
	for _, addition := range additions {
		if addition != "" && !slices.Contains(values, addition) {
			values = append(values, addition)
		}
	}
	return values
}

func sortedUnique(values []string) []string {
	values = appendUnique(nil, values...)
	slices.Sort(values)
	return values
}

func classify(value Value) (decision, resolution, reason string, topSuccess bool) {
	switch value.State {
	case StateExact:
		return "PASS", "EXACT", "ALL_OBSERVATIONS_EXACT", true
	case StateDirectUnknown:
		return "UNKNOWN", "LOWER_RESOLUTION", "DIRECT_UNKNOWN_PRESERVED", false
	case StateDependencyBlocked:
		return "UNKNOWN", "LOWER_RESOLUTION", "DEPENDENCY_BLOCKED_PRESERVED", false
	case StateInvariantOnly:
		return "HOLD", "INVARIANT_ONLY", "KNOWN_INVARIANT_PRESERVED", false
	case StateMixedUnresolved:
		return "UNKNOWN", "LOWER_RESOLUTION", "MIXED_UNRESOLVED_PRESERVED", false
	default:
		return "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_COMPOSITION_STATE", false
	}
}

func transitionState(state State) string {
	if state == StateExact {
		return "DISCHARGED"
	}
	return "OPEN"
}

func deriveState(value Evidence) State {
	if !value.ObservedAvailable {
		if value.Dependency != nil && (value.Dependency.State == "OPEN" || value.Dependency.State == "UNKNOWN") {
			return StateDependencyBlocked
		}
		return StateDirectUnknown
	}
	if value.InvariantEvidence != "" {
		return StateInvariantOnly
	}
	return StateExact
}
