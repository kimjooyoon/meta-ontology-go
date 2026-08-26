package valueexecution

func allIndicatorsSatisfied(indicators []Indicator) bool {
	if len(indicators) != 16 {
		return false
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
	}
	return true
}
