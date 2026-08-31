package artifactemit

import "sort"

type symbolicReaderSpec struct {
	audience, sourceResolution, effectiveResolution string
}

var symbolicReaderSpecs = []symbolicReaderSpec{
	{"USER", "USER_VISIBLE", "DECISION_AND_COUNTS_ONLY"},
	{"TOOL_AUTHOR", "TOOL_CONTRACT", "INDICATOR_CONTRACT_ONLY"},
	{"GOVERNOR", "FULL_RECEIPT", "SOURCE_BOUND_RECEIPT_ONLY"},
}

func symbolicReaderBuildViews(
	source SymbolicValueReachability,
	checks *symbolicReaderChecks,
) []SymbolicValueReaderProjectionView {
	views := make(map[string][]SymbolicValueContractView)
	for _, view := range source.Views {
		views[view.Audience] = append(views[view.Audience], view)
	}
	ids := make(map[string][]string)
	satisfied := make(map[string]int)
	for _, indicator := range source.Indicators {
		for _, audience := range indicator.Audiences {
			ids[audience] = append(ids[audience], indicator.ID)
			if indicator.Satisfied {
				satisfied[audience]++
			}
		}
	}
	projections := make([]SymbolicValueReaderProjectionView, 0, len(symbolicReaderSpecs))
	allResolutions := true
	for _, spec := range symbolicReaderSpecs {
		sort.Strings(ids[spec.audience])
		present := len(views[spec.audience]) == 1
		bound := present && symbolicReaderViewBound(
			views[spec.audience][0], satisfied[spec.audience], len(ids[spec.audience]),
		)
		canonical := present && views[spec.audience][0].Resolution == spec.sourceResolution
		allResolutions = allResolutions && canonical
		symbolicReaderSetChecks(checks, spec.audience, present, bound)
		projections = append(projections, SymbolicValueReaderProjectionView{
			Audience: spec.audience, SourceResolution: symbolicReaderSourceResolution(views[spec.audience]),
			EffectiveResolution: spec.effectiveResolution, IndicatorIDs: ids[spec.audience],
			Coordinates: symbolicReaderCoordinates(satisfied[spec.audience], len(ids[spec.audience])),
		})
	}
	checks.ReaderResolutions = allResolutions
	checks.UserNested = symbolicReaderSubset(ids["USER"], ids["TOOL_AUTHOR"])
	checks.ToolNested = symbolicReaderSubset(ids["TOOL_AUTHOR"], ids["GOVERNOR"])
	return projections
}

func symbolicReaderSourceResolution(views []SymbolicValueContractView) string {
	if len(views) != 1 {
		return "UNKNOWN"
	}
	return views[0].Resolution
}
