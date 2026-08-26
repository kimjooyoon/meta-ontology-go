package valueexecution

func allIndicatorsSatisfied(indicators []Indicator) bool {
	if len(indicators) != ValueIndicatorCount {
		return false
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
	}
	return true
}

func firstUnsatisfiedIndicator(indicators []Indicator) string {
	if len(indicators) != ValueIndicatorCount {
		return "INDICATOR_DENOMINATOR_MISMATCH"
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return indicator.ID
		}
	}
	return ""
}
