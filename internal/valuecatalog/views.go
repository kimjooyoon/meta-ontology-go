package valuecatalog

func buildViews(indicators []Indicator) []View {
	return []View{
		buildView("USER", "CATALOG_OUTCOME", indicators, []int{4, 6, 9}),
		buildView("TOOL_AUTHOR", "CATALOG_CONTRACT", indicators, []int{0, 1, 2, 3, 4, 5, 6, 7, 9, 10}),
		buildView("GOVERNOR", "FULL_RECEIPT", indicators, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
	}
}

func buildView(audience, resolution string, indicators []Indicator, indexes []int) View {
	view := View{Audience: audience, Resolution: resolution, Total: len(indexes)}
	for _, index := range indexes {
		indicator := indicators[index]
		view.IndicatorIDs = append(view.IndicatorIDs, indicator.ID)
		if indicator.Satisfied {
			view.Satisfied++
		}
	}
	view.BasisPoints = coordinate(view.Satisfied, view.Total).BasisPoints
	return view
}
