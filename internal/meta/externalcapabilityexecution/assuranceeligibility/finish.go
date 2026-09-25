package assuranceeligibility

func finish(report Report) Report {
	report.Indicators = buildIndicators(report)
	counts := map[string]*[2]int{"DRIVER": {}, "OUTCOME": {}, "GUARDRAIL": {}}
	proofs := map[string]*[2]int{"FOUNDATION": {}, "COHERENCE": {}, "REGRESSION": {}}
	for _, value := range report.Indicators {
		count := counts[value.Class]
		count[1]++
		proof := proofs[value.ProofChoice]
		proof[1]++
		if value.Satisfied {
			count[0]++
			proof[0]++
			report.Summary.IndicatorCompleted++
		}
	}
	report.Summary.IndicatorTotal = len(report.Indicators)
	report.Summary.IndicatorCoverageBPS = report.Summary.IndicatorCompleted * 10000 / len(report.Indicators)
	report.Summary.DriverCompleted, report.Summary.DriverTotal = counts["DRIVER"][0], counts["DRIVER"][1]
	report.Summary.OutcomeCompleted, report.Summary.OutcomeTotal = counts["OUTCOME"][0], counts["OUTCOME"][1]
	report.Summary.GuardrailCompleted, report.Summary.GuardrailTotal = counts["GUARDRAIL"][0], counts["GUARDRAIL"][1]
	for _, choice := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		value := proofs[choice]
		status := "UNSATISFIED"
		if value[0] == value[1] {
			status = "SATISFIED"
		}
		report.Proofs = append(report.Proofs, Proof{Choice: choice, Status: status,
			Satisfied: value[0], Total: value[1]})
	}
	report.RepositoryWrites = report.Summary.RepositoryWrites + report.Summary.ExternalRepositoryWrites
	report.OfficialMutationCount, report.PromotionApplied = report.Summary.OfficialMutations, report.Summary.Promotions
	return sealReport(report)
}
