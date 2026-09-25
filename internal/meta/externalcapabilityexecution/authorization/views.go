package authorization

func makeReaderViews(indicators []Indicator) []ReaderView {
	local := makeView("LOCAL_EVIDENCE", indicators, true)
	authorization := makeView("AUTHORIZATION", indicators, false)
	return []ReaderView{local, authorization}
}

func makeView(reader string, indicators []Indicator, excludeFoundation bool) ReaderView {
	view := ReaderView{Reader: reader, Resolution: ResolutionExact}
	for _, indicator := range indicators {
		if excludeFoundation && indicator.MetricID ==
			"gooo.metric.external-capability-authorization-policy-foundation.v1" {
			continue
		}
		view.Total++
		if indicator.Status == StatusSatisfied {
			view.Completed++
		}
		if indicator.Status == StatusUnknown {
			view.Resolution = ResolutionUnknown
		}
	}
	view.BasisPoints = view.Completed * 10000 / view.Total
	return view
}
