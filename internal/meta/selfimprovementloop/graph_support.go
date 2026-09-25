package selfimprovementloop

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func decisionRank(decision string) int {
	switch decision {
	case DecisionClosed:
		return 0
	case DecisionUnknown:
		return 1
	case DecisionRefuted:
		return 2
	default:
		return 3
	}
}

// Prioritize makes known refutation outrank unresolved evidence.
func Prioritize(decisions ...string) string {
	best := DecisionClosed
	for _, decision := range decisions {
		if decisionRank(decision) > decisionRank(best) {
			best = decision
		}
	}
	return best
}
