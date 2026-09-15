package artifactemit

func symbolicReaderSetChecks(
	checks *symbolicReaderChecks,
	audience string,
	present, bound bool,
) {
	switch audience {
	case "USER":
		checks.UserPresent, checks.UserCountBound = present, bound
	case "TOOL_AUTHOR":
		checks.ToolPresent, checks.ToolCountBound = present, bound
	case "GOVERNOR":
		checks.GovernorPresent, checks.GovernorCountBound = present, bound
	}
}

func symbolicReaderViewBound(
	view SymbolicValueContractView,
	satisfied, total int,
) bool {
	coordinates := symbolicReaderCoordinates(satisfied, total)
	return view.Satisfied == coordinates.Satisfied &&
		view.Total == coordinates.Total &&
		view.BasisPoints == coordinates.BasisPoints
}

func symbolicReaderSubset(left, right []string) bool {
	set := make(map[string]struct{}, len(right))
	for _, value := range right {
		set[value] = struct{}{}
	}
	for _, value := range left {
		if _, found := set[value]; !found {
			return false
		}
	}
	return true
}

func symbolicReaderLowerResolution(readers []SymbolicValueReaderProjectionView) {
	for index := range readers {
		readers[index].EffectiveResolution = "INVARIANT_ONLY"
	}
}
