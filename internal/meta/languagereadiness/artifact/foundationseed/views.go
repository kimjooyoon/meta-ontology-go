package foundationseed

func views(values []Indicator) []View {
	toolIDs := make([]string, 0, len(values)-1)
	governorIDs := make([]string, 0, len(values))
	for _, value := range values {
		governorIDs = append(governorIDs, value.ID)
		if value.ID != "authority-denied" {
			toolIDs = append(toolIDs, value.ID)
		}
	}
	return []View{
		view("USER", "FOUNDATION_OUTCOME", values, []string{
			"seed-scope-exact", "readiness-delta-claims-zero", "authority-denied",
		}),
		view("TOOL_AUTHOR", "EXHAUSTION_CONTRACT", values, toolIDs),
		view("GOVERNOR", "FULL_RECEIPT", values, governorIDs),
	}
}

func view(audience, resolution string, values []Indicator, ids []string) View {
	passed := make(map[string]bool, len(values))
	for _, value := range values {
		passed[value.ID] = value.Passed
	}
	result := View{Audience: audience, Resolution: resolution,
		Total: len(ids), IndicatorIDs: append([]string(nil), ids...)}
	for _, id := range ids {
		if passed[id] {
			result.Satisfied++
		}
	}
	result.BasisPoints = result.Satisfied * 10000 / result.Total
	return result
}
