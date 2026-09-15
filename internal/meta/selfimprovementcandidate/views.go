package selfimprovementcandidate

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

func buildViews(indicators []Indicator) []View {
	user := []int{6, 8, 9, 14, 15}
	tool := make([]int, 12)
	for index := range tool {
		tool[index] = index
	}
	governor := make([]int, len(indicators))
	for index := range governor {
		governor[index] = index
	}
	return []View{
		buildView("USER", "USER_VISIBLE", indicators, user),
		buildView("TOOL_AUTHOR", "TOOL_CONTRACT", indicators, tool),
		buildView("GOVERNOR", "FULL_RECEIPT", indicators, governor),
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
