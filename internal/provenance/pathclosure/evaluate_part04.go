package pathclosure

func sortResultIDs(result *Result) {
	sortIDs(result.Required)
	sortIDs(result.Complete)
	sortIDs(result.Missing)
	sortIDs(result.Malformed)
	sortIDs(result.Duplicate)
}
