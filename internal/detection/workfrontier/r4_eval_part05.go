package workfrontier

func r4UnknownWithResult(result R4Result, reason string) R4Result {
	result.Status = R4StatusUnknown
	result.Reason = reason
	result.Quality = R4StatusUnknown
	result.FullSuiteRequired = true
	result.Selected = nil
	result.SelectedIDs = nil
	result.WorkIDs = nil
	return normalizeR4Result(result)
}
func r4FailClosed(graph r4Graph, reason string) R4Result {
	result := r4Unknown(graph, reason)
	result.Status = R4StatusFailClosed
	result.Quality = R4StatusFailClosed
	return result
}
