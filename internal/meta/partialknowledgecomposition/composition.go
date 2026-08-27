package partialknowledgecomposition

import "slices"

// Compose combines knowledge causes without allowing a less precise input to
// become a more precise output. The mixed state keeps both unresolved causes.
func Compose(left, right Operand) Value {
	value := Value{Contributors: []string{left.Operation, right.Operation}}
	for _, operand := range []Operand{left, right} {
		switch operand.State {
		case StateDirectUnknown:
			value.DirectUnknowns = appendUnique(value.DirectUnknowns, operand.Operation)
		case StateDependencyBlocked:
			value.BlockedDependencies = appendUnique(value.BlockedDependencies, operand.BlockedDependency)
		case StateInvariantOnly:
			value.PreservedInvariants = appendUnique(value.PreservedInvariants, operand.Invariants...)
		}
	}
	value.DirectUnknowns = sortedUnique(value.DirectUnknowns)
	value.BlockedDependencies = sortedUnique(value.BlockedDependencies)
	value.PreservedInvariants = sortedUnique(value.PreservedInvariants)
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

func classify(value Value) (string, string, bool) {
	switch value.State {
	case StateExact:
		return "PASS", "ALL_OPERATIONS_EXACT", true
	case StateDirectUnknown:
		return "FAIL_CLOSED", "DIRECT_UNKNOWN_NOT_PROMOTED", false
	case StateDependencyBlocked:
		return "FAIL_CLOSED", "DEPENDENCY_BLOCKED_NOT_PROMOTED", false
	case StateInvariantOnly:
		return "HOLD", "KNOWN_INVARIANT_PRESERVED", false
	case StateMixedUnresolved:
		return "FAIL_CLOSED", "MIXED_UNRESOLVED_KNOWLEDGE", false
	default:
		return "FAIL_CLOSED", "UNKNOWN_COMPOSITION_STATE", false
	}
}
