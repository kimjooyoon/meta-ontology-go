package verify

import "sort"

const (
	exact             = "EXACT"
	directUnknown     = "DIRECT_UNKNOWN"
	dependencyBlocked = "DEPENDENCY_BLOCKED"
	invariantOnly     = "INVARIANT_ONLY"
	mixedUnresolved   = "MIXED_UNRESOLVED"
)

func compose(left, right operand) value {
	result := value{Contributors: []string{left.Operation, right.Operation}}
	for _, current := range []operand{left, right} {
		switch current.State {
		case directUnknown:
			result.DirectUnknowns = unique(result.DirectUnknowns, current.Operation)
		case dependencyBlocked:
			result.BlockedDependencies = unique(result.BlockedDependencies, current.BlockedDependency)
		case invariantOnly:
			result.PreservedInvariants = unique(result.PreservedInvariants, current.Invariants...)
		}
	}
	sort.Strings(result.DirectUnknowns)
	sort.Strings(result.BlockedDependencies)
	sort.Strings(result.PreservedInvariants)
	sort.Strings(result.Contributors)
	switch {
	case len(result.DirectUnknowns) != 0 && len(result.BlockedDependencies) != 0:
		result.State = mixedUnresolved
	case len(result.DirectUnknowns) != 0:
		result.State = directUnknown
	case len(result.BlockedDependencies) != 0:
		result.State = dependencyBlocked
	case len(result.PreservedInvariants) != 0:
		result.State = invariantOnly
	default:
		result.State = exact
	}
	return result
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

func classify(value value) (string, string, bool) {
	switch value.State {
	case exact:
		return "PASS", "ALL_OPERATIONS_EXACT", true
	case directUnknown:
		return "FAIL_CLOSED", "DIRECT_UNKNOWN_NOT_PROMOTED", false
	case dependencyBlocked:
		return "FAIL_CLOSED", "DEPENDENCY_BLOCKED_NOT_PROMOTED", false
	case invariantOnly:
		return "HOLD", "KNOWN_INVARIANT_PRESERVED", false
	case mixedUnresolved:
		return "FAIL_CLOSED", "MIXED_UNRESOLVED_KNOWLEDGE", false
	default:
		return "FAIL_CLOSED", "UNKNOWN_COMPOSITION_STATE", false
	}
}

func transitionState(state string) string {
	switch state {
	case exact:
		return "DISCHARGED"
	case directUnknown:
		return "UNKNOWN"
	case dependencyBlocked:
		return "BLOCKED"
	case invariantOnly:
		return "INVARIANT_PRESERVED"
	default:
		return "UNRESOLVED"
	}
}
