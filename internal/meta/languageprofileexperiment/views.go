package languageprofileexperiment

func buildViews(indicators []Indicator) []View {
	return []View{
		view("USER", "USER_VISIBLE", indicators, []string{
			"profile.receipts", "profile.samples", "profile.successful-executions",
			"resource.wall-observations", "resource.allocation-observations"}),
		view("TOOL_AUTHOR", "TOOL_CONTRACT", indicators, indicatorIDs(indicators[:10])),
		view("GOVERNOR", "FULL_RECEIPT", indicators, indicatorIDs(indicators)),
	}
}

func view(audience, resolution string, indicators []Indicator, ids []string) View {
	satisfied := 0
	for _, id := range ids {
		for _, item := range indicators {
			if item.ID == id && item.Satisfied {
				satisfied++
			}
		}
	}
	basis := 0
	if len(ids) > 0 {
		basis = satisfied * 10000 / len(ids)
	}
	return View{Audience: audience, Resolution: resolution, Satisfied: satisfied,
		Total: len(ids), BasisPoints: basis, IndicatorIDs: append([]string(nil), ids...)}
}

func indicatorIDs(indicators []Indicator) []string {
	ids := make([]string, len(indicators))
	for index, item := range indicators {
		ids[index] = item.ID
	}
	return ids
}
