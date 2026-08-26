package artifactemit

func symbolicReaderMetricViews(
	indicators []SymbolicValueContractIndicator,
) []SymbolicValueContractView {
	specs := []struct {
		audience, resolution string
	}{
		{"USER", "DECISION_AND_COUNTS_ONLY"},
		{"TOOL_AUTHOR", "INDICATOR_CONTRACT_ONLY"},
		{"GOVERNOR", "FULL_RECEIPT"},
	}
	result := make([]SymbolicValueContractView, 0, len(specs))
	for _, spec := range specs {
		satisfied, total := 0, 0
		for _, indicator := range indicators {
			if symbolicReaderHasAudience(indicator.Audiences, spec.audience) {
				total++
				if indicator.Satisfied {
					satisfied++
				}
			}
		}
		coordinates := symbolicReaderCoordinates(satisfied, total)
		result = append(result, SymbolicValueContractView{
			Audience: spec.audience, Resolution: spec.resolution,
			Satisfied: coordinates.Satisfied, Total: coordinates.Total,
			BasisPoints: coordinates.BasisPoints,
		})
	}
	return result
}

func symbolicReaderHasAudience(audiences []string, expected string) bool {
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}
