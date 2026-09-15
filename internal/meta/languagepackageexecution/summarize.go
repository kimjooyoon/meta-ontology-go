package languagepackageexecution

func summarize(contract Contract, evidence []CaseEvidence, results []CaseResult) Summary {
	summary := Summary{CasesSatisfied: countSatisfied(results), CasesTotal: len(contract.Cases)}
	for _, item := range evidence {
		summary.RepositoryWrites += item.Receipt.Effects.RepositoryWrites
		summary.MutationAuthorities += boolInt(item.Receipt.Effects.MutationAuthority)
	}
	for _, result := range results {
		if result.Reason == "PACKAGE_EXECUTION_DECISION_UNKNOWN" || result.Resolution == "LOWER_RESOLUTION" {
			summary.UnknownDecisions++
		}
		if result.Satisfied && result.Decision == "FAIL_CLOSED" {
			summary.DiagnosticRejections++
		}
	}
	if len(evidence) > 0 && len(results) > 0 && results[0].Satisfied {
		summary.SourceFilesLoaded = len(evidence[0].Receipt.Sources)
		summary.PackageExecutions = 1
		summary.EventsObserved = len(evidence[0].Receipt.Events)
	}
	if len(results) > 1 {
		summary.DeterministicReplays = boolInt(results[1].Satisfied)
	}
	return summary
}

func countSatisfied(values []CaseResult) int {
	count := 0
	for _, value := range values {
		count += boolInt(value.Satisfied)
	}
	return count
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
