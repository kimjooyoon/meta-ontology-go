package toolchainusecases

func finish(report Report) Report {
	summary := Summary{Total: totalCases, RepositoryWrites: report.RepositoryWrites}
	for _, item := range report.Cases {
		switch item.Status {
		case "SATISFIED":
			summary.Satisfied++
		case "UNRESOLVED":
			summary.Unresolved++
		default:
			summary.NotSatisfied++
		}
		if item.Status == "SATISFIED" && item.ObservedDecision == DecisionPass {
			summary.PassPaths++
		}
		if item.Status == "SATISFIED" && item.ObservedDecision == DecisionClosed {
			summary.FailClosedPaths++
		}
	}
	summary.Executed = summary.Total - summary.Unresolved
	summary.ReadinessBPS = summary.Satisfied * 10_000 / totalCases
	report.Summary = summary
	registryDrift := boolInt(report.Source.RegistryDigest != registryDigest())
	ready := summary.Satisfied == totalCases && summary.Unresolved == 0 &&
		report.RepositoryWrites == 0 && report.Source.ConceptRepositoryWrites == 0 &&
		!report.MutationAuthorized && registryDrift == 0
	report.Decision, report.Reason, report.Resolution = DecisionClosed, "USE_CASE_MISMATCH", ResolutionExact
	if summary.Unresolved > 0 {
		report.Reason, report.Resolution = "USE_CASE_EVIDENCE_UNKNOWN", ResolutionLower
	} else if ready {
		report.Decision, report.Reason = DecisionPass, "EXECUTABLE_USE_CASES_PROVEN"
	}
	report.Indicators = indicators(report, registryDrift)
	report.Proofs = proofs(report, registryDrift)
	return seal(report)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
